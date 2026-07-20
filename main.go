package main

import (
	"os"
	"time"

	"github.com/killmonday/fscanx/Plugins"
	"github.com/killmonday/fscanx/common"
	"github.com/killmonday/fscanx/tui"
	//"net/http"
	_ "net/http/pprof"
)

func main() {
	//go func() {
	//	http.ListenAndServe("localhost:6060", nil)
	//}()

	// Interactive TUI mode: only on a capable TTY (Unix), or when the user
	// explicitly forces it with -tui. This preserves the Win7 / CI /
	// masscan|fscanx pipe paths (which fall back to the CLI).
	forceTUI := false
	for _, a := range os.Args[1:] {
		if a == "-tui" || a == "--tui" {
			forceTUI = true
			break
		}
	}
	if tui.IsInteractive(forceTUI) {
		args, err := tui.RunTUI()
		if err != nil {
			// Cancelled or failed to start TUI — bail out cleanly.
			if err.Error() != "cancelled" {
				os.Exit(1)
			}
			os.Exit(0)
		}
		// Rebuild os.Args so the existing flag-driven flow consumes them.
		os.Args = append([]string{"fscanx"}, args...)
	}

	var Info common.HostInfo
	start := time.Now()
	common.Flag(&Info)
	common.Parse(&Info)

	// 检查是否有 --std 参数（通过flag）
	if common.ScanWithStdInput {
		Plugins.ScanFromStdin()
		common.LogSuccess("[*] scan done! cost: %s\n", time.Since(start))
		common.LogWG.Wait() //等待所有日志打印和写入文件等等事件
		close(common.Results)
		return
	}
	Plugins.Scan(Info)
	common.LogSuccess("[*] scan done! cost: %s\n", time.Since(start))
	common.LogWG.Wait() //等待所有日志打印和写入文件等等事件
	close(common.Results)
}
