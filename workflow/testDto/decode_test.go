package testDto

import (
	"encoding/json"
	"fmt"
	"testing"
	"tinyflow/workflow/model"
)

var postData = "{\"name\":\"请假\",\"nameSpace\":\"A公司\",\"resource\":{\"name\":\"发起人\",\"type\":\"start\",\"nodeId\":\"sid-startevent\",\"childNode\":{\"type\":\"route\",\"prevId\":\"sid-startevent\",\"nodeId\":\"8b5c_debb\",\"conditionNodes\":[{\"name\":\"条件1\",\"type\":\"condition\",\"prevId\":\"8b5c_debb\",\"nodeId\":\"da89_be76\",\"properties\":{\"conditions\":[[{\"type\":\"dingtalk_actioner_value_condition\",\"paramKey\":\"DDHolidayField-J2BWEN12__options\",\"paramLabel\":\"请假类型\",\"paramValue\":\"\",\"paramValues\":[\"年假\"],\"oriValue\":[\"年假\",\"事假\",\"病假\",\"调休\",\"产假\",\"婚假\",\"例假\",\"丧假\"],\"isEmpty\":false}]]},\"childNode\":{\"name\":\"UNKNOWN\",\"type\":\"approver\",\"prevId\":\"da89_be76\",\"nodeId\":\"735c_0854\",\"properties\":{\"activateType\":\"ONE_BY_ONE\",\"agreeAll\":false,\"actionerRules\":[{\"type\":\"target_management\",\"level\":1,\"isEmpty\":false,\"autoUp\":true}]}}},{\"name\":\"条件2\",\"type\":\"condition\",\"prevId\":\"8b5c_debb\",\"nodeId\":\"a97f_9517\",\"properties\":{\"conditions\":[[{\"type\":\"dingtalk_actioner_value_condition\",\"paramKey\":\"DDHolidayField-J2BWEN12__options\",\"paramLabel\":\"请假类型\",\"paramValue\":\"\",\"paramValues\":[\"调休\"],\"oriValue\":[\"年假\",\"事假\",\"病假\",\"调休\",\"产假\",\"婚假\",\"例假\",\"丧假\"],\"isEmpty\":false}]]},\"childNode\":{\"name\":\"UNKNOWN\",\"type\":\"approver\",\"prevId\":\"a97f_9517\",\"nodeId\":\"5891_395b\",\"properties\":{\"activateType\":\"ALL\",\"agreeAll\":true,\"actionerRules\":[{\"type\":\"target_label\",\"labelNames\":\"财务\",\"labels\":427529103,\"isEmpty\":false,\"memberCount\":2,\"actType\":\"and\"}],\"noneActionerAction\":\"auto\"}}}],\"properties\":{},\"childNode\":{\"name\":\"UNKNOWN\",\"type\":\"approver\",\"prevId\":\"8b5c_debb\",\"nodeId\":\"59ba_8815\",\"properties\":{\"activateType\":\"ALL\",\"agreeAll\":true,\"actionerRules\":[{\"type\":\"target_label\",\"labelNames\":\"人事\",\"labels\":427529104,\"isEmpty\":false,\"memberCount\":1,\"actType\":\"and\"}]}}}}}"

func TestDecode(t *testing.T) {
	def := model.ProcessDefine{}
	err := json.Unmarshal([]byte(postData), &def)
	if err != nil {
		return
	}
	fmt.Println(def)
	fmt.Println(def.Nodes.IfConfigValid())
	flow, err := def.Nodes.GenExecFlow(&map[string]string{"DDHolidayField-J2BWEN12__options": "年假"})
	if err != nil {
		return
	}
	fmt.Println(flow)
}
