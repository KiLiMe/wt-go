wt-go 便携模式

exe 旁边有这个 data 目录, wt-go 就是便携版: 它的文件全部写在 data 里,
不碰 %LOCALAPPDATA%\wt-go。把这个目录删掉, wt-go 回到安装版的行为,
文件放回 %LOCALAPPDATA%\wt-go, 和安装在机器上的那份共用。

data 里的东西由 wt-go 自己建:

  config\   settings.json、window-state.json、locale(上次选的语言)
  logs\     wt-go.log, 上一次的是 wt-go.log.1
  crashes\  崩溃时写的 crash-<时间>.log, 没崩过就是空的

换台机器、或者挪到别的文件夹: 把整个文件夹一起复制过去就行。
