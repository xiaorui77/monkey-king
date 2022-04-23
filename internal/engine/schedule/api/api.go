package api

type TaskManage interface {
	SetProcess(domain string, num int)
	DeleteTask(domain string, id uint64) bool
	GetTree(domain string) interface{}
}
