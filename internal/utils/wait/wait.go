package wait

import "time"

// WaitWhen 每隔一定时间去检查fun指定的函数状态, 直到为true, 否则一直等待(阻塞)
func WaitWhen(fun func() bool) {
	for {
		if fun() {
			return
		}
		time.Sleep(time.Second)
	}
}
