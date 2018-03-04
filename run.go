package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	errorFormat = "%+v"
	progName    = "DockerTinyproxy"
	proxyConf   = "/etc/tinyproxy/tinyproxy.conf"
	tailLog     = "/var/log/tinyproxy/tinyproxy.log"
)

var (
	logger *zap.SugaredLogger
)

func init() {
	viper.AutomaticEnv()
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		exitOnFailure(err)
	}
	logger = zapLogger.Sugar()
}
func main() {
	defer logger.Sync()

	// Start script
	logger.Infof("%q script started...", progName)

	// Stop Tinyproxy if running
	logger.Infof("Checking for running Tinyproxy service...")
	execCmds([][2]string{
		{`[ "$(pidof tail)" ] && killall tail`, "tail tinyproxy log stopping"},
		{`[ "$(pidof tinyproxy)" ] && killall tinyproxy`, "tinyproxy service stopping"},
	})

	// Set ACL in Tinyproxy config

	execCmds([][2]string{
		{fmt.Sprintf(`sed -i -e"/^Allow.*/d" %s`, proxyConf), "remove Allow lines from tinyproxy conf"},
	})
	allows := make([]string, 0)
	for _, allow := range strings.Split(strings.Trim(viper.GetString(`TINYPROXY_ALLOW`), " "), " ") {
		if allow == "" {
			continue
		}
		logger.Infof("TINYPROXY_ALLOW: %q", allow)
		allows = append(allows, fmt.Sprintf("Allow %s\n", allow))
	}
	f, err := os.OpenFile(proxyConf, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Errorf("open the tinyproxy conf: "+errorFormat, err)
	}
	defer f.Close()
	_, err = f.WriteString(strings.Join(allows, ""))
	if err != nil {
		logger.Errorf("write to the tinyproxy conf: "+errorFormat, err)
	}
	err = f.Close()
	if err != nil {
		logger.Errorf("close the tinyproxy conf: "+errorFormat, err)
	}

	// Enable log to file
	execCmds([][2]string{
		{fmt.Sprintf(`touch %s`, tailLog), ""},
		{fmt.Sprintf(`sed -i -e"s,^#LogFile,LogFile," %s`, proxyConf), ""},
	})

	// Start Tinyproxy
	logger.Infof("Starting Tinyproxy service...")
	tinyproxy := exec.Command("sh", "-c", "/usr/sbin/tinyproxy")
	err = tinyproxy.Start()
	if err != nil {
		logger.Infof("tinyproxy starting: "+errorFormat, err)
	}

	// Tail Tinyproxy log
	logger.Infof("Tailing Tinyproxy log...")
	tail := exec.Command("sh", "-c", fmt.Sprintf("tail -f %s", tailLog))
	tail.Stdout = os.Stdout
	tail.Stderr = os.Stderr
	err = tail.Run()
	if err != nil {
		logger.Infof("tail -f %s: "+errorFormat, tailLog, err)
	}

	// End
	err = tinyproxy.Wait()
	if err != nil {
		logger.Infof("tinyproxy ending: "+errorFormat, err)
	}
	logger.Infof("%q script ended.", progName)
}

// exitOnFailure prints a fatal error message and exits the process with status 1.
func exitOnFailure(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "[CRIT] "+errorFormat+"\n", err)
	os.Exit(1)
}

func execCmds(cmds [][2]string) {
	for _, command := range cmds {
		out, err := exec.Command("sh", "-c", command[0]).Output()
		if command[1] != "" {
			logger.Infof("%s", command[1])
		}
		if len(out) > 0 {
			logger.Infof("%s", out)
		}
		if err != nil {
			logger.Infof("%s %q: "+errorFormat, command[1], command[0], err)
		}
	}
}
