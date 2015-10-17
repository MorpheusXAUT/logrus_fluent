package logrus_fluent

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/fluent/fluent-logger-golang/fluent"
)

const (
	TagName      = "fluent"
	TagField     = "tag"
	MessageField = "message"
)

var defaultLevels = []logrus.Level{
	logrus.PanicLevel,
	logrus.FatalLevel,
	logrus.ErrorLevel,
	logrus.WarnLevel,
	logrus.InfoLevel,
}

type fluentHook struct {
	host        string
	port        int
	application string
	levels      []logrus.Level
}

func NewHook(host string, port int, application string) *fluentHook {
	return &fluentHook{
		host:        host,
		port:        port,
		application: application,
		levels:      defaultLevels,
	}
}

func getTagAndDel(entry *logrus.Entry, application string) string {
	var v interface{}
	var ok bool
	if v, ok = entry.Data[TagField]; !ok {
		return fmt.Sprintf("%s.%s", application, entry.Level)
	}

	var val string
	if val, ok = v.(string); !ok {
		return fmt.Sprintf("%s.%s", application, entry.Level)
	}
	delete(entry.Data, TagField)
	return val
}

func setLevelString(entry *logrus.Entry) {
	entry.Data["level"] = entry.Level.String()
}

func setMessage(entry *logrus.Entry) {
	if _, ok := entry.Data[MessageField]; !ok {
		entry.Data[MessageField] = entry.Message
	}
}

func setCaller(entry *logrus.Entry, skip int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "???"
		line = 0
	} else {
		lastSlash := strings.LastIndex(file, "/")
		if lastSlash >= 0 {
			folderSlash := strings.LastIndex(file[:lastSlash], "/")
			if folderSlash >= 0 {
				file = file[folderSlash+1:]
			} else {
				file = file[lastSlash+1:]
			}
		}
	}

	entry.Data["caller"] = fmt.Sprintf("%s:%d", file, line)
}

func (hook *fluentHook) Fire(entry *logrus.Entry) error {
	logger, err := fluent.New(fluent.Config{
		FluentHost: hook.host,
		FluentPort: hook.port,
	})
	if err != nil {
		return err
	}
	defer logger.Close()

	setLevelString(entry)
	tag := getTagAndDel(entry, hook.application)
	if tag != entry.Message {
		setMessage(entry)
	}

	setCaller(entry, 5)

	data := ConvertToValue(entry.Data, TagName)
	err = logger.PostWithTime(tag, entry.Time, data)
	return err
}

func (hook *fluentHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *fluentHook) SetLevels(levels []logrus.Level) {
	hook.levels = levels
}
