package internal

import (
	"fmt"
	"log"
	"os"

	"github.com/mo7amed4522/tayara_api/config"
	"github.com/mo7amed4522/tayara_api/global"
	"gorm.io/gorm/logger"
)

type Writer struct {
	config config.GeneralDB
	writer logger.Writer
}

func NewWriter(config config.GeneralDB) *Writer {
	return &Writer{
		config: config,
		writer: log.New(os.Stdout, "\r\n", log.LstdFlags),
	}
}

// Printf Formatted print log
func (c *Writer) Printf(message string, data ...any) {
	if c.config.LogZap {
		switch c.config.LogLevel() {
		case logger.Silent:
			global.GVA_LOG.Debug(fmt.Sprintf(message, data...))
		case logger.Error:
			global.GVA_LOG.Error(fmt.Sprintf(message, data...))
		case logger.Warn:
			global.GVA_LOG.Warn(fmt.Sprintf(message, data...))
		case logger.Info:
			global.GVA_LOG.Info(fmt.Sprintf(message, data...))
		default:
			global.GVA_LOG.Info(fmt.Sprintf(message, data...))
		}
		return
	}
	if c.writer == nil {
		log.Printf(message, data...)
		return
	}
	c.writer.Printf(message, data...)
}
