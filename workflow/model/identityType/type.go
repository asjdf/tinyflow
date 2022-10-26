package identityType

type Type int

const (
	// CANDIDATE 候选人
	CANDIDATE Type = iota
	// PARTICIPANT 参与人
	PARTICIPANT
	// MANAGER 上级领导
	MANAGER
	// NOTIFIER 抄送人
	NOTIFIER
)

var ToStr = map[Type]string{CANDIDATE: "候选", PARTICIPANT: "参与", MANAGER: "主管", NOTIFIER: "抄送"}
