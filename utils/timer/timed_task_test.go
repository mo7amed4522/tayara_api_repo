package timer

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var job = mockJob{}

type mockJob struct{}

func (job mockJob) Run() {
	mockFunc()
}

func mockFunc() {
	time.Sleep(time.Second)
	fmt.Println("1s...")
}

func TestNewTimerTask(t *testing.T) {
	tm := NewTimerTask()
	_tm := tm.(*timer)

	{
		_, err := _tm.AddTaskByFunc("func", "@every 1s", mockFunc, "Test mockfunc")
		assert.Nil(t, err)
		_, ok := _tm.cronList["func"]
		if !ok {
			t.Error("no find func")
		}
	}

	{
		_, err := _tm.AddTaskByJob("job", "@every 1s", job, "Test job mockfunc")
		assert.Nil(t, err)
		_, ok := _tm.cronList["job"]
		if !ok {
			t.Error("no find job")
		}
	}

	{
		_, ok := _tm.FindCron("func")
		if !ok {
			t.Error("no find func")
		}
		_, ok = _tm.FindCron("job")
		if !ok {
			t.Error("no find job")
		}
		_, ok = _tm.FindCron("none")
		if ok {
			t.Error("find none")
		}
	}
	{
		_tm.Clear("func")
		_, ok := _tm.FindCron("func")
		if ok {
			t.Error("find func")
		}
	}
	{
		a := _tm.FindCronList()
		b, c := _tm.FindCron("job")
		fmt.Println(a, b, c)
	}
}
