package enum

type TicketStatus int
type TicketTag string

const (
	TicketStatusNormal TicketStatus = iota //0:未售，1：已售，2：已退，3:已删除
	TicketStatusSold
	TicketStatusRefund
	TicketStatusDeleted
)

// 简化为枚举表示车次，前端可以按照枚举值查询或是修改车次并显示为详细信息
const (
	G101 TicketTag = "G101" //8-9点左右 成都到河南
	G102 TicketTag = "G102" //12-13点左右 河南到成都
	G103 TicketTag = "G103" //18-19点左右 河南到杭州
)

func (s TicketStatus) String() string {
	switch s {
	case TicketStatusNormal:
		return "未售"
	case TicketStatusSold:
		return "已售"
	case TicketStatusDeleted:
		return "已删除"
	default:
		return "UNKNOWN"
	}
}

func (s TicketTag) String() string {
	switch s {
	case G101:
		return "G101"
	case G102:
		return "G102"
	case G103:
		return "G103"
	default:
		return "UNKNOWN"
	}
}
