package cache

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

type fakeRedisEntry struct {
	value     string
	expiresAt time.Time
}

type fakeRedis struct {
	mu      sync.Mutex
	entries map[string]fakeRedisEntry
}

func newFakeRedisClient(t *testing.T) redis.UniversalClient {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &fakeRedis{entries: make(map[string]fakeRedisEntry)}
	go server.serve(listener)

	client := redis.NewClient(&redis.Options{
		Addr:            listener.Addr().String(),
		Protocol:        2,
		DisableIdentity: true,
	})
	t.Cleanup(func() {
		_ = client.Close()
		_ = listener.Close()
	})
	return client
}

func (s *fakeRedis) serve(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go s.handle(conn)
	}
}

func (s *fakeRedis) handle(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	for {
		command, err := readRESPArray(reader)
		if err != nil {
			return
		}
		if err := s.execute(writer, command); err != nil {
			return
		}
		if err := writer.Flush(); err != nil {
			return
		}
	}
}

func readRESPArray(reader *bufio.Reader) ([]string, error) {
	line, err := readRESPLine(reader)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(line, "*") {
		return nil, errors.New("expected RESP array")
	}
	length, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, err
	}
	args := make([]string, length)
	for i := range args {
		line, err = readRESPLine(reader)
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(line, "$") {
			return nil, errors.New("expected RESP bulk string")
		}
		size, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, err
		}
		payload := make([]byte, size+2)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return nil, err
		}
		args[i] = string(payload[:size])
	}
	return args, nil
}

func readRESPLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"), nil
}

func (s *fakeRedis) execute(writer *bufio.Writer, args []string) error {
	if len(args) == 0 {
		return errors.New("empty Redis command")
	}
	switch strings.ToUpper(args[0]) {
	case "HELLO":
		_, err := writer.WriteString("*14\r\n" +
			"$6\r\nserver\r\n$5\r\nredis\r\n" +
			"$7\r\nversion\r\n$5\r\n7.0.0\r\n" +
			"$5\r\nproto\r\n:2\r\n" +
			"$2\r\nid\r\n:1\r\n" +
			"$4\r\nmode\r\n$10\r\nstandalone\r\n" +
			"$4\r\nrole\r\n$6\r\nmaster\r\n" +
			"$7\r\nmodules\r\n*0\r\n")
		return err
	case "GET":
		s.mu.Lock()
		entry, found := s.currentEntry(args[1])
		s.mu.Unlock()
		if !found {
			_, err := writer.WriteString("$-1\r\n")
			return err
		}
		_, err := fmt.Fprintf(writer, "$%d\r\n%s\r\n", len(entry.value), entry.value)
		return err
	case "SET":
		if len(args) < 3 {
			return errors.New("invalid SET command")
		}
		var expiry time.Duration
		nx := false
		for i := 3; i < len(args); i++ {
			switch strings.ToUpper(args[i]) {
			case "PX":
				i++
				milliseconds, err := strconv.ParseInt(args[i], 10, 64)
				if err != nil {
					return err
				}
				expiry = time.Duration(milliseconds) * time.Millisecond
			case "NX":
				nx = true
			}
		}
		s.mu.Lock()
		_, exists := s.currentEntry(args[1])
		if nx && exists {
			s.mu.Unlock()
			_, err := writer.WriteString("$-1\r\n")
			return err
		}
		entry := fakeRedisEntry{value: args[2]}
		if expiry > 0 {
			entry.expiresAt = time.Now().Add(expiry)
		}
		s.entries[args[1]] = entry
		s.mu.Unlock()
		_, err := writer.WriteString("+OK\r\n")
		return err
	case "PTTL":
		s.mu.Lock()
		entry, found := s.currentEntry(args[1])
		s.mu.Unlock()
		if !found {
			_, err := writer.WriteString(":-2\r\n")
			return err
		}
		if entry.expiresAt.IsZero() {
			_, err := writer.WriteString(":-1\r\n")
			return err
		}
		_, err := fmt.Fprintf(writer, ":%d\r\n", time.Until(entry.expiresAt).Milliseconds())
		return err
	case "DEL":
		s.mu.Lock()
		_, existed := s.currentEntry(args[1])
		delete(s.entries, args[1])
		s.mu.Unlock()
		if existed {
			_, err := writer.WriteString(":1\r\n")
			return err
		}
		_, err := writer.WriteString(":0\r\n")
		return err
	case "EVALSHA":
		_, err := writer.WriteString("-NOSCRIPT No matching script. Please use EVAL.\r\n")
		return err
	case "EVAL":
		if len(args) < 5 {
			return errors.New("invalid EVAL command")
		}
		key := args[3]
		token := args[4]
		s.mu.Lock()
		entry, found := s.currentEntry(key)
		deleted := 0
		if found && entry.value == token {
			delete(s.entries, key)
			deleted = 1
		}
		s.mu.Unlock()
		_, err := fmt.Fprintf(writer, ":%d\r\n", deleted)
		return err
	default:
		return fmt.Errorf("unsupported Redis command %q", args[0])
	}
}

func (s *fakeRedis) currentEntry(key string) (fakeRedisEntry, bool) {
	entry, found := s.entries[key]
	if found && !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		delete(s.entries, key)
		return fakeRedisEntry{}, false
	}
	return entry, found
}
