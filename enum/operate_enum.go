package enum

type OperateType int

const (
	OperateOK     OperateType = 200 //操作成功
	OperateFailed OperateType = 500 //操作失败
)

func (o OperateType) String() string {
	switch o {
	case OperateOK:
		return "操作成功"
	case OperateFailed:
		return "操作失败"
	}
	return "UNKNOWN"
}
