package model

import (
	"container/list"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"tinyflow/utils"
)

// ActionConditionType 条件类型
type ActionConditionType int

const (
	// RANGE 条件类型: 范围
	RANGE ActionConditionType = iota
	// VALUE 条件类型： 值
	VALUE
)

// ActionConditionTypes 所有条件类型
var ActionConditionTypes = [...]string{RANGE: "dingtalk_actioner_range_condition", VALUE: "dingtalk_actioner_value_condition"}

// ActionRuleType 审批人类型
type ActionRuleType int

const (
	MANAGER ActionRuleType = iota
	LABEL
)

var ActionRuleTypes = [...]string{MANAGER: "target_management", LABEL: "target_label"}

type NodeType int

const (
	// START 类型start
	START NodeType = iota
	ROUTE
	CONDITION
	APPROVER
	NOTIFIER
)

var NodeTypes = [...]string{START: "start", ROUTE: "route", CONDITION: "condition", APPROVER: "approver", NOTIFIER: "notifier"}

// Node 工作流中的单个节点
type Node struct {
	Name           string          `json:"name,omitempty"`
	Type           string          `json:"type,omitempty"`
	NodeID         string          `json:"nodeId,omitempty"`
	PrevID         string          `json:"prevId,omitempty"`
	ChildNode      *Node           `json:"childNode,omitempty"`
	ConditionNodes []*Node         `json:"conditionNodes,omitempty"`
	Properties     *NodeProperties `json:"properties,omitempty"`
}

func (n *Node) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	err := json.Unmarshal(bytes, n)
	return err
}
func (n *Node) Value() (driver.Value, error) {
	return json.Marshal(n)
}

type NodeProperties struct {
	// ONE_BY_ONE 代表依次审批
	ActivateType       string             `json:"activateType,omitempty"`
	AgreeAll           bool               `json:"agreeAll,omitempty"`
	Conditions         [][]*NodeCondition `json:"conditions,omitempty"`
	ActionerRules      []*ActionerRule    `json:"actionerRules,omitempty"`
	NoneActionerAction string             `json:"noneActionerAction,omitempty"`
}

type NodeCondition struct {
	Type       string `json:"type,omitempty"`
	ParamKey   string `json:"paramKey,omitempty"`
	ParamLabel string `json:"paramLabel,omitempty"`
	IsEmpty    bool   `json:"isEmpty,omitempty"`
	// 类型为range
	LowerBound      string `json:"lowerBound,omitempty"`
	LowerBoundEqual string `json:"lowerBoundEqual,omitempty"`
	UpperBoundEqual string `json:"upperBoundEqual,omitempty"`
	UpperBound      string `json:"upperBound,omitempty"`
	BoundEqual      string `json:"boundEqual,omitempty"`
	Unit            string `json:"unit,omitempty"`
	// 类型为 value
	ParamValues []string    `json:"paramValues,omitempty"`
	OriValue    []string    `json:"oriValue,omitempty"`
	Conds       []*NodeCond `json:"conds,omitempty"`
}

type ActionerRule struct {
	Type       string `json:"type,omitempty"`
	LabelNames string `json:"labelNames,omitempty"`
	Labels     int    `json:"labels,omitempty"`
	IsEmpty    bool   `json:"isEmpty,omitempty"`
	// 表示需要通过的人数 如果是会签
	MemberCount int8 `json:"memberCount,omitempty"`
	// and 表示会签 or表示或签，默认为或签
	ActType string `json:"actType,omitempty"`
	Level   int8   `json:"level,omitempty"`
	AutoUp  bool   `json:"autoUp,omitempty"`
}

type NodeCond struct {
	Type  string    `json:"type,omitempty"`
	Value string    `json:"value,omitempty"`
	Attrs *NodeUser `json:"attrs,omitempty"`
}

type NodeUser struct {
	Name   string `json:"name,omitempty"`
	Avatar string `json:"avatar,omitempty"`
}

// NodeInfo 节点信息 用于生成执行流
type NodeInfo struct {
	NodeID      string `json:"nodeId"`
	Type        string `json:"type"`
	Aprover     string `json:"approver"`
	AproverType string `json:"aproverType"`
	MemberCount int8   `json:"memberCount"`
	Level       int8   `json:"level"`
	ActType     string `json:"actType"`
}

type ProcessInputVariable *map[string]string

// IfConfigValid 检查Node的配置是否合法
func (n *Node) IfConfigValid() error {
	// 节点名称是否有效
	if len(n.NodeID) == 0 {
		return errors.New("节点的【nodeId】不能为空！！")
	}
	// 检查类型是否有效
	if len(n.Type) == 0 {
		return errors.New("节点【" + n.NodeID + "】的类型【type】不能为空")
	}
	isValidType := false
	for _, val := range NodeTypes {
		if val == n.Type {
			isValidType = true
			break
		}
	}
	if !isValidType {
		str, _ := utils.ToJSONStr(NodeTypes)
		return errors.New("节点【" + n.NodeID + "】的类型为【" + n.Type + "】，为无效类型,有效类型为" + str)
	}
	// 检查审批节点是否有审批人 消息节点是否有配置
	if n.Type == NodeTypes[APPROVER] || n.Type == NodeTypes[NOTIFIER] {
		if n.Properties == nil || n.Properties.ActionerRules == nil {
			return errors.New("节点【" + n.NodeID + "】的Properties属性不能为空，如：`\"properties\": {\"actionerRules\": [{\"type\": \"target_label\",\"labelNames\": \"人事\",\"memberCount\": 1,\"actType\": \"and\"}],}`")
		}
	}
	// 条件节点是否存在
	if n.ConditionNodes != nil { // 存在条件节点
		if len(n.ConditionNodes) == 1 {
			return errors.New("节点【" + n.NodeID + "】条件节点下的节点数必须大于1")
		}
		// 根据条件变量选择节点索引
		err := CheckConditionNode(n.ConditionNodes)
		if err != nil {
			return err
		}
	}

	// 子节点是否存在
	if n.ChildNode != nil {
		return n.ChildNode.IfConfigValid()
	}
	return nil
}

// CheckConditionNode 检查条件节点
func CheckConditionNode(nodes []*Node) error {
	for _, node := range nodes {
		if node.Properties == nil {
			return errors.New("节点【" + node.NodeID + "】的Properties对象为空值！！")
		}
		if len(node.Properties.Conditions) == 0 {
			return errors.New("节点【" + node.NodeID + "】的Conditions对象为空值！！")
		}
		err := node.IfConfigValid()
		if err != nil {
			return err
		}
	}
	return nil
}

// GenExecFlow 根据传入的变量提前规划执行流
func (n *Node) GenExecFlow(variable ProcessInputVariable) (*list.List, error) {
	l := list.New()
	err := parseProcessConfig(n, variable, l)
	return l, err
}

func parseProcessConfig(node *Node, variable ProcessInputVariable, list *list.List) (err error) {
	switch node.Type {
	case NodeTypes[APPROVER], NodeTypes[NOTIFIER]:
		var aprover string
		if node.Properties.ActionerRules[0].Type == ActionRuleTypes[MANAGER] {
			aprover = "主管"
		} else {
			aprover = node.Properties.ActionerRules[0].LabelNames
		}
		list.PushBack(NodeInfo{
			NodeID:      node.NodeID,
			Type:        node.Properties.ActionerRules[0].Type,
			Aprover:     aprover,
			AproverType: node.Type,
			MemberCount: node.Properties.ActionerRules[0].MemberCount,
			ActType:     node.Properties.ActionerRules[0].ActType,
		})
		break
	default:
	}

	// 存在条件节点
	if node.ConditionNodes != nil {
		// 如果条件节点只有一个或者条件只有一个，直接返回第一个
		if variable == nil || len(node.ConditionNodes) == 1 {
			err = parseProcessConfig(node.ConditionNodes[0].ChildNode, variable, list)
			if err != nil {
				return err
			}
		} else {
			// 根据条件变量选择节点索引
			condNode, err := GetConditionNode(node.ConditionNodes, variable)
			if err != nil {
				return err
			}
			if condNode == nil {
				str, _ := utils.ToJSONStr(variable)
				return errors.New("节点【" + node.NodeID + "】找不到符合条件的子节点,检查变量【var】值是否匹配," + str)
				// panic(err)
			}
			err = parseProcessConfig(condNode, variable, list)
			if err != nil {
				return err
			}

		}
	}
	// 存在子节点
	if node.ChildNode != nil {
		err = parseProcessConfig(node.ChildNode, variable, list)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetConditionNode 获取条件节点
func GetConditionNode(nodes []*Node, maps ProcessInputVariable) (result *Node, err error) {
	map2 := *maps
	for _, node := range nodes {
		var flag int
		for _, v := range node.Properties.Conditions[0] {
			paramValue := map2[v.ParamKey]
			if len(paramValue) == 0 {
				return nil, errors.New("流程启动变量【var】的key【" + v.ParamKey + "】的值不能为空")
			}
			yes, err := checkConditions(v, paramValue)
			if err != nil {
				return nil, err
			}
			if yes {
				flag++
			}
		}
		// 满足所有条件
		if flag == len(node.Properties.Conditions[0]) {
			result = node
		}
	}
	return result, nil
}

func checkConditions(cond *NodeCondition, value string) (bool, error) {
	// 判断类型
	switch cond.Type {
	case ActionConditionTypes[RANGE]:
		val, err := strconv.Atoi(value)
		if err != nil {
			return false, err
		}
		if len(cond.LowerBound) == 0 && len(cond.UpperBound) == 0 && len(cond.LowerBoundEqual) == 0 && len(cond.UpperBoundEqual) == 0 && len(cond.BoundEqual) == 0 {
			return false, errors.New("条件【" + cond.Type + "】的上限或者下限值不能全为空")
		}
		// 判断下限，lowerBound
		if len(cond.LowerBound) > 0 {
			low, err := strconv.Atoi(cond.LowerBound)
			if err != nil {
				return false, err
			}
			if val <= low {
				// fmt.Printf("val:%d小于lowerBound:%d\n", val, low)
				return false, nil
			}
		}
		if len(cond.LowerBoundEqual) > 0 {
			le, err := strconv.Atoi(cond.LowerBoundEqual)
			if err != nil {
				return false, err
			}
			if val < le {
				// fmt.Printf("val:%d小于lowerBound:%d\n", val, low)
				return false, nil
			}
		}
		// 判断上限,upperBound包含等于
		if len(cond.UpperBound) > 0 {
			upper, err := strconv.Atoi(cond.UpperBound)
			if err != nil {
				return false, err
			}
			if val >= upper {
				return false, nil
			}
		}
		if len(cond.UpperBoundEqual) > 0 {
			ge, err := strconv.Atoi(cond.UpperBoundEqual)
			if err != nil {
				return false, err
			}
			if val > ge {
				return false, nil
			}
		}
		if len(cond.BoundEqual) > 0 {
			equal, err := strconv.Atoi(cond.BoundEqual)
			if err != nil {
				return false, err
			}
			if val != equal {
				return false, nil
			}
		}
		return true, nil
	case ActionConditionTypes[VALUE]:
		if len(cond.ParamValues) == 0 {
			return false, errors.New("条件节点【" + cond.Type + "】的 【paramValues】数组不能为空，值如：'paramValues:['调休','年假']")
		}
		for _, val := range cond.ParamValues {
			if value == val {
				return true, nil
			}
		}
		// log.Printf("key:" + cond.ParamKey + "找不到对应的值")
		return false, nil
	default:
		str, _ := utils.ToJSONStr(ActionConditionTypes)
		return false, errors.New("未知的NodeCondition类型【" + cond.Type + "】,正确类型应为以下中的一个:" + str)
	}
}
