##pprof   
###安装go-torch 需要几个依赖库  
github.com\brendangregg\FlameGraph  
github.com\jessevdk\go-flags  
github.com\fatih\color   
###针对windows环境 github.com\uber\go-torch\renderer\flamegraph.go   
flameGraph := findInPath(flameGraphScripts)   
if flameGraph == "" {   
	return nil, errNoPerlScript   
}   
if runtime.GOOS == "windows" {   
	return runScript("perl", append([]string{flameGraph}, args...), graphInput)  
}   
  return runScript(flameGraph, args, graphInput)  
}   
修改完成后 需要重新安装  
###使用方法:  
方法一:go-torch http://127.0.0.1:6060/debug/pprof/profile -f profile001.svg      
方法二:go-torch profile.prof profile.svg   
###go tool pprof使用方法  
go tool pprof [执行程序][分析文件] 
linux下进入终端后可gif结果保存图片显示 svg
