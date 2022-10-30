package identityType

type Type int

const (
	// CANDIDATE 当前审批人或审批用户组
	CANDIDATE Type = iota
	// PARTICIPANT 参与人 标记已参与某审核
	PARTICIPANT
	// MANAGER 上级领导
	MANAGER
	// NOTIFIER 抄送人
	NOTIFIER
)

var ToStr = map[Type]string{CANDIDATE: "审批人员", PARTICIPANT: "参与", MANAGER: "主管", NOTIFIER: "抄送"}
