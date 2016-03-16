package log

import (
	"fmt"
	"github.com/cihub/seelog"
	"goserver/config"
)

var Logger seelog.LoggerInterface

func init() {
}

func SetupLoggerFromConfig(c *config.Config) {
	logConfigTmp := `
		<seelog minlevel="%s">
			<outputs formatid="common">
				<rollingfile type="size" filename="%s" maxsize="%d" maxrolls="%d" />
				<filter levels="error">
					<file path="%s" formatid="error"/>
				</filter>
			</outputs>
			<formats>
				<format id="common" format="%s" />
				<format id="error" format="%s" />
			</formats>
		</seelog>
	`
	logConfig := fmt.Sprintf(logConfigTmp,
		c.Logging.Level,
		c.Logging.Filename,
		c.Logging.Maxsize,
		c.Logging.Maxrolls,
		c.Logging.ErrorFilename,
		"%Date/%Time [%LEV] %Msg%n",
		"%Date/%Time %File %FullPath %Func %Msg%n")
	logger, err := seelog.LoggerFromConfigAsBytes([]byte(logConfig))
	if err != nil {
		panic(err)
	}
	//seelog.ReplaceLogger(logger)
	seelog.Current.Flush()
	seelog.Current.Close()
	seelog.Current = logger
	Logger = logger
}

func Flush() { seelog.Flush() }

func Critical(msg ...interface{}) { seelog.Critical(msg) }
func Fatal(msg ...interface{})    { seelog.Critical(msg) }
func Error(msg ...interface{})    { seelog.Error(msg) }
func Warn(msg ...interface{})     { seelog.Warn(msg) }
func Info(msg ...interface{})     { seelog.Info(msg) }
func Debug(msg ...interface{})    { seelog.Debug(msg) }

func Criticalf(msg string, vals ...interface{}) { seelog.Criticalf(msg, vals...) }
func Fatalf(msg string, vals ...interface{})    { seelog.Criticalf(msg, vals...) }
func Errorf(msg string, vals ...interface{})    { seelog.Errorf(msg, vals...) }
func Warnf(msg string, vals ...interface{})     { seelog.Warnf(msg, vals...) }
func Infof(msg string, vals ...interface{})     { seelog.Infof(msg, vals...) }
func Debugf(msg string, vals ...interface{})    { seelog.Debugf(msg, vals...) }

/*
import (
	"dockercenter/config"
	"github.com/cloudfoundry/gosteno"
	"os"
)

var logger *gosteno.Logger

func init() {
	stenoConfig := &gosteno.Config{
		Sinks: []gosteno.Sink{gosteno.NewIOSink(os.Stderr)},
		Codec: gosteno.NewJsonCodec(),
		Level: gosteno.LOG_ALL,
	}

	gosteno.Init(stenoConfig)
	logger = gosteno.NewLogger("dockercenter")
}

func SetupLoggerFromConfig(c *config.Config) {
	l, err := gosteno.GetLogLevel(c.Logging.Level)
	if err != nil {
		panic(err)
	}

	s := make([]gosteno.Sink, 0)
	if c.Logging.File != "" {
		s = append(s, gosteno.NewFileSink(c.Logging.File))
	} else {
		s = append(s, gosteno.NewIOSink(os.Stdout))
	}

	if c.Logging.Syslog != "" {
		s = append(s, gosteno.NewSyslogSink(c.Logging.Syslog))
	}

	stenoConfig := &gosteno.Config{
		Sinks: s,
		Codec: gosteno.NewJsonCodec(),
		Level: l,
	}

	gosteno.Init(stenoConfig)
	logger = gosteno.NewLogger("dockercenter")
}

func Fatal(msg string) { logger.Fatal(msg) }
func Error(msg string) { logger.Error(msg) }
func Warn(msg string)  { logger.Warn(msg) }
func Info(msg string)  { logger.Info(msg) }
func Debug(msg string) { logger.Debug(msg) }

func Fatald(data map[string]interface{}, msg string) { logger.Fatald(data, msg) }
func Errord(data map[string]interface{}, msg string) { logger.Errord(data, msg) }
func Warnd(data map[string]interface{}, msg string)  { logger.Warnd(data, msg) }
func Infod(data map[string]interface{}, msg string)  { logger.Infod(data, msg) }
func Debugd(data map[string]interface{}, msg string) { logger.Debugd(data, msg) }

func Fatalf(msg string, vals ...interface{}) { logger.Fatalf(msg, vals...) }
func Errorf(msg string, vals ...interface{}) { logger.Errorf(msg, vals...) }
func Warnf(msg string, vals ...interface{})  { logger.Warnf(msg, vals...) }
func Infof(msg string, vals ...interface{})  { logger.Infof(msg, vals...) }
func Debugf(msg string, vals ...interface{}) { logger.Debugf(msg, vals...) }
*/
