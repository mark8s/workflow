package bll

import (
	"flow/model"
	"flow/schema"
	"flow/util"
	"fmt"
	"sync"
	"time"
)

// Flow 流程管理
type Flow struct {
	sync.RWMutex
	FlowModel *model.Flow `inject:""`
}

// GetFlow 获取流程数据
func (a *Flow) GetFlow(recordID string) (*schema.Flow, error) {
	return a.FlowModel.GetFlow(recordID)
}

// GetFlowByCode 根据编号查询流程数据
func (a *Flow) GetFlowByCode(code string) (*schema.Flow, error) {
	return a.FlowModel.GetFlowByCode(code)
}

// QueryFlowByCode 根据流程编号查询流程数据
func (a *Flow) QueryFlowByCode(flowCode string) ([]*schema.Flow, error) {
	return a.FlowModel.QueryFlowByCode(flowCode)
}

// CreateFlow 创建流程数据
func (a *Flow) CreateFlow(flow *schema.Flow, nodes *schema.NodeOperating, forms *schema.FormOperating) error {
	if flow.Flag == 0 {
		flow.Flag = 1
	}
	return a.FlowModel.CreateFlow(flow, nodes, forms)
}

// GetNode 获取流程节点
func (a *Flow) GetNode(recordID string) (*schema.Node, error) {
	return a.FlowModel.GetNode(recordID)
}

// GetFlowInstance 获取流程实例
func (a *Flow) GetFlowInstance(recordID string) (*schema.FlowInstance, error) {
	return a.FlowModel.GetFlowInstance(recordID)
}

// GetFlowInstanceByNode 根据节点实例获取流程实例
func (a *Flow) GetFlowInstanceByNode(nodeInstanceID string) (*schema.FlowInstance, error) {
	return a.FlowModel.GetFlowInstanceByNode(nodeInstanceID)
}

// GetNodeInstance 获取流程节点实例
func (a *Flow) GetNodeInstance(recordID string) (*schema.NodeInstance, error) {
	return a.FlowModel.GetNodeInstance(recordID)
}

// QueryNodeRouters 查询节点路由
func (a *Flow) QueryNodeRouters(sourceNodeID string) ([]*schema.NodeRouter, error) {
	return a.FlowModel.QueryNodeRouters(sourceNodeID)
}

// QueryNodeAssignments 查询节点指派
func (a *Flow) QueryNodeAssignments(nodeID string) ([]*schema.NodeAssignment, error) {
	return a.FlowModel.QueryNodeAssignments(nodeID)
}

// CreateNodeInstance 创建节点实例
func (a *Flow) CreateNodeInstance(flowInstanceID, nodeID string, inputData []byte, candidates []string) (string, error) {
	nodeInstance := &schema.NodeInstance{
		RecordID:       util.UUID(),
		FlowInstanceID: flowInstanceID,
		NodeID:         nodeID,
		InputData:      string(inputData),
		Status:         1,
		Created:        time.Now().Unix(),
	}

	var nodeCandidates []*schema.NodeCandidate
	for _, c := range candidates {
		nodeCandidates = append(nodeCandidates, &schema.NodeCandidate{
			RecordID:       util.UUID(),
			NodeInstanceID: nodeInstance.RecordID,
			CandidateID:    c,
			Created:        nodeInstance.Created,
		})
	}

	err := a.FlowModel.CreateNodeInstance(nodeInstance, nodeCandidates)
	if err != nil {
		return "", err
	}

	return nodeInstance.RecordID, nil
}

// DoneNodeInstance 完成节点实例
func (a *Flow) DoneNodeInstance(nodeInstanceID, processor string, outData []byte) error {
	// 加锁保证节点实例的处理过程
	a.Lock()
	defer a.Unlock()

	nodeInstance, err := a.FlowModel.GetNodeInstance(nodeInstanceID)
	if err != nil {
		return err
	} else if nodeInstance == nil || nodeInstance.Status == 2 {
		return fmt.Errorf("无效的处理节点")
	}

	info := map[string]interface{}{
		"processor":    processor,
		"process_time": time.Now().Unix(),
		"out_data":     string(outData),
		"status":       2,
		"updated":      time.Now().Unix(),
	}
	return a.FlowModel.UpdateNodeInstance(nodeInstanceID, info)
}

// CheckFlowInstanceTodo 检查流程实例待办事项
func (a *Flow) CheckFlowInstanceTodo(flowInstanceID string) (bool, error) {
	return a.FlowModel.CheckFlowInstanceTodo(flowInstanceID)
}

// DoneFlowInstance 完成流程实例
func (a *Flow) DoneFlowInstance(flowInstanceID string) error {
	info := map[string]interface{}{
		"status": 9,
	}
	return a.FlowModel.UpdateFlowInstance(flowInstanceID, info)
}

// StopFlowInstance 停止流程实例
func (a *Flow) StopFlowInstance(flowInstanceID string) error {
	info := map[string]interface{}{
		"status": 9,
	}
	return a.FlowModel.UpdateFlowInstance(flowInstanceID, info)
}

// LaunchFlowInstance2 发起流程实例（基于流程ID），返回流程实例、开始事件节点实例
func (a *Flow) LaunchFlowInstance2(flowID, userID string, status int, inputData []byte) (*schema.FlowInstance, *schema.NodeInstance, error) {
	node, err := a.GetNodeByFlowAndTypeCode(flowID, "startEvent")
	if err != nil {
		return nil, nil, err
	} else if node == nil {
		return nil, nil, fmt.Errorf("未知的流程节点")
	}

	flowInstance := &schema.FlowInstance{
		RecordID:   util.UUID(),
		FlowID:     flowID,
		Launcher:   userID,
		LaunchTime: time.Now().Unix(),
		Status:     int64(status),
		Created:    time.Now().Unix(),
	}

	nodeInstance := &schema.NodeInstance{
		RecordID:       util.UUID(),
		FlowInstanceID: flowInstance.RecordID,
		NodeID:         node.RecordID,
		InputData:      string(inputData),
		Status:         1,
		Created:        flowInstance.Created,
	}

	err = a.FlowModel.CreateFlowInstance(flowInstance, nodeInstance)
	if err != nil {
		return nil, nil, err
	}

	return flowInstance, nodeInstance, nil
}

// LaunchFlowInstance 发起流程实例
func (a *Flow) LaunchFlowInstance(flowCode, nodeCode, launcher string, inputData []byte) (*schema.NodeInstance, error) {
	flow, err := a.FlowModel.GetFlowByCode(flowCode)
	if err != nil {
		return nil, err
	} else if flow == nil {
		return nil, nil
	}

	node, err := a.FlowModel.GetNodeByCode(flow.RecordID, nodeCode)
	if err != nil {
		return nil, err
	} else if node == nil {
		return nil, nil
	}

	flowInstance := &schema.FlowInstance{
		RecordID:   util.UUID(),
		FlowID:     flow.RecordID,
		Launcher:   launcher,
		LaunchTime: time.Now().Unix(),
		Status:     1,
		Created:    time.Now().Unix(),
	}

	nodeInstance := &schema.NodeInstance{
		RecordID:       util.UUID(),
		FlowInstanceID: flowInstance.RecordID,
		NodeID:         node.RecordID,
		InputData:      string(inputData),
		Status:         1,
		Created:        flowInstance.Created,
	}

	err = a.FlowModel.CreateFlowInstance(flowInstance, nodeInstance)
	if err != nil {
		return nil, err
	}

	return nodeInstance, nil
}

// QueryNodeCandidates 查询节点候选人
func (a *Flow) QueryNodeCandidates(nodeInstanceID string) ([]*schema.NodeCandidate, error) {
	return a.FlowModel.QueryNodeCandidates(nodeInstanceID)
}

// CheckNodeCandidate 检查节点候选人
func (a *Flow) CheckNodeCandidate(nodeInstanceID, userID string) (bool, error) {
	return a.FlowModel.CheckNodeCandidate(nodeInstanceID, userID)
}

// QueryTodo 查询用户的待办节点实例数据
func (a *Flow) QueryTodo(typeCode, flowCode, userID string, count int) ([]*schema.FlowTodoResult, error) {
	return a.FlowModel.QueryTodo(typeCode, flowCode, userID, count)
}

// GetTodoByID 根据ID获取待办
func (a *Flow) GetTodoByID(nodeInstanceID string) (*schema.FlowTodoResult, error) {
	return a.FlowModel.GetTodoByID(nodeInstanceID)
}

// GetDoneByID 根据ID获取已办
func (a *Flow) GetDoneByID(nodeInstanceID string) (*schema.FlowDoneResult, error) {
	return a.FlowModel.GetDoneByID(nodeInstanceID)
}

// QueryDone 查询用户的已办数据
func (a *Flow) QueryDone(typeCode, flowCode, userID string, lastTime int64, count int) ([]*schema.FlowDoneResult, error) {
	return a.FlowModel.QueryDone(typeCode, flowCode, userID, lastTime, count)
}

// GetDoneCount 获取已办数量
func (a *Flow) GetDoneCount(userID string) (int64, error) {
	return a.FlowModel.GetDoneCount(userID)
}

// QueryAllFlowPage 查询流程分页数据
func (a *Flow) QueryAllFlowPage(params schema.FlowQueryParam, pageIndex, pageSize uint) (int64, []*schema.FlowQueryResult, error) {
	return a.FlowModel.QueryAllFlowPage(params, pageIndex, pageSize)
}

// DeleteFlow 删除流程
func (a *Flow) DeleteFlow(flowID string) error {
	return a.FlowModel.DeleteFlow(flowID)
}

// QueryHistory 查询流程实例历史数据
func (a *Flow) QueryHistory(flowInstanceID string) ([]*schema.FlowHistoryResult, error) {
	return a.FlowModel.QueryHistory(flowInstanceID)
}

// QueryDoneIDs 查询已办理的流程实例ID列表
func (a *Flow) QueryDoneIDs(flowCode, userID string) ([]string, error) {
	return a.FlowModel.QueryDoneIDs(flowCode, userID)
}

// QueryGroupFlowPage 查询流程分组分页数据
func (a *Flow) QueryGroupFlowPage(params schema.FlowQueryParam, pageIndex, pageSize uint) (int64, []*schema.FlowQueryResult, error) {
	return a.FlowModel.QueryGroupFlowPage(params, pageIndex, pageSize)
}

// UpdateFlowInfo 更新流程
func (a *Flow) UpdateFlowInfo(recordID string, info map[string]interface{}) error {
	return a.FlowModel.Update(recordID, info)
}

// UpdateFlowStatus 更新流程状态
func (a *Flow) UpdateFlowStatus(recordID string, status int) error {
	info := map[string]interface{}{
		"updated": time.Now().Unix(),
		"status":  status,
	}

	return a.UpdateFlowInfo(recordID, info)
}

// QueryFlowVersion 查询流程版本数据
func (a *Flow) QueryFlowVersion(recordID string) ([]*schema.FlowQueryResult, error) {
	flow, err := a.FlowModel.GetFlow(recordID)
	if err != nil {
		return nil, err
	} else if flow == nil {
		return nil, nil
	}

	return a.FlowModel.QueryFlowVersion(flow.Code)
}

// QueryFlowIDsByType 根据类型查询流程ID列表
func (a *Flow) QueryFlowIDsByType(typeCodes ...string) ([]string, error) {
	return a.FlowModel.QueryFlowIDsByType(typeCodes...)
}

// QueryFlowByIDs 根据流程ID查询流程数据
func (a *Flow) QueryFlowByIDs(flowIDs []string) ([]*schema.FlowQueryResult, error) {
	return a.FlowModel.QueryFlowByIDs(flowIDs)
}

// GetFlowFormByNodeID 获取流程节点表单
func (a *Flow) GetFlowFormByNodeID(nodeID string) (*schema.Form, error) {
	return a.FlowModel.GetFlowFormByNodeID(nodeID)
}

// QueryNodeByTypeCodeAndFlowIDs 根据节点类型和流程ID列表查询节点数据
func (a *Flow) QueryNodeByTypeCodeAndFlowIDs(typeCode string, flowIDs ...string) ([]*schema.Node, error) {
	return a.FlowModel.QueryNodeByTypeCodeAndFlowIDs(typeCode, flowIDs...)
}

// GetNodeByFlowAndTypeCode 根据流程ID和节点类型获取节点数据
func (a *Flow) GetNodeByFlowAndTypeCode(flowID, typeCode string) (*schema.Node, error) {
	return a.FlowModel.GetNodeByFlowAndTypeCode(flowID, typeCode)
}

// GetForm 获取流程表单
func (a *Flow) GetForm(formID string) (*schema.Form, error) {
	return a.FlowModel.GetForm(formID)
}

// GetNodeProperty 获取节点属性
func (a *Flow) GetNodeProperty(nodeID string) (map[string]string, error) {
	items, err := a.FlowModel.QueryNodeProperty(nodeID)
	if err != nil {
		return nil, err
	}

	data := make(map[string]string)
	for _, item := range items {
		data[item.Name] = item.Value
	}
	return data, nil
}

// CreateNodeTiming 创建定时节点
func (a *Flow) CreateNodeTiming(item *schema.NodeTiming) error {
	item.ID = 0
	return a.FlowModel.CreateNodeTiming(item)
}

// DeleteNodeTiming 删除定时节点
func (a *Flow) DeleteNodeTiming(nodeInstanceID string) error {
	return a.FlowModel.UpdateNodeTiming(nodeInstanceID, map[string]interface{}{"deleted": time.Now().Unix()})
}

// QueryExpiredNodeTiming 查询到期的定时节点
func (a *Flow) QueryExpiredNodeTiming() ([]*schema.NodeTiming, error) {
	return a.FlowModel.QueryExpiredNodeTiming()
}

// QueryLaunchFlowInstanceResult 查询发起的流程实例数据
func (a *Flow) QueryLaunchFlowInstanceResult(launcher, typeCode, flowCode string, lastID int64, count int) ([]*schema.FlowInstanceResult, error) {
	return a.FlowModel.QueryLaunchFlowInstanceResult(launcher, typeCode, flowCode, lastID, count)
}

// QueryTodoFlowInstanceResult 查询待办的流程实例数据
func (a *Flow) QueryTodoFlowInstanceResult(userID, typeCode, flowCode string, lastID int64, count int) ([]*schema.FlowInstanceResult, error) {
	return a.FlowModel.QueryTodoFlowInstanceResult(userID, typeCode, flowCode, lastID, count)
}

// QueryWebTodoFlowInstanceResult web查询待办的流程实例数据
func (a *Flow) QueryWebTodoFlowInstanceResult(userID, typeCode, flowCode string, count int, ParamSearchList map[string]string) ([]*schema.FlowWebInstanceResult, int64, error) {
	return a.FlowModel.QueryTodoWebFlowInstanceResult(userID, typeCode, flowCode, count, ParamSearchList)
}

// QueryHandleFlowInstanceResult 查询处理的流程实例结果
func (a *Flow) QueryHandleFlowInstanceResult(processor, typeCode, flowCode string, lastID int64, count int) ([]*schema.FlowInstanceResult, error) {
	return a.FlowModel.QueryHandleFlowInstanceResult(processor, typeCode, flowCode, lastID, count)
}

// QueryWebHandleFlowInstanceResult web查询处理的流程实例结果
func (a *Flow) QueryWebHandleFlowInstanceResult(processor, typeCode, flowCode string, lastID int64, count int, ParamSearchList map[string]string) ([]*schema.FlowInstanceResult, int64, error) {
	return a.FlowModel.QueryWebHandleFlowInstanceResult(processor, typeCode, flowCode, lastID, count, ParamSearchList)
}

// QueryLastNodeInstances 查询流程实例的最后一个节点实例
func (a *Flow) QueryLastNodeInstances(flowInstanceIDs []string) (map[string]*schema.NodeInstance, error) {
	items, err := a.FlowModel.QueryLastNodeInstances(flowInstanceIDs)
	if err != nil {
		return nil, err
	}

	data := make(map[string]*schema.NodeInstance)
	for _, item := range items {
		data[item.FlowInstanceID] = item
	}
	return data, nil
}

// QueryWebLastNodeInstances web查询流程实例的最后一个节点实例
func (a *Flow) QueryWebLastNodeInstances(flowInstanceIDs []string, ParamSearchList map[string]string, isComplete bool) (map[string]*schema.NodeInstance, error) {

	items, err := a.FlowModel.QueryWebLastNodeInstances(flowInstanceIDs, ParamSearchList, isComplete)
	if err != nil {
		return nil, err
	}

	data := make(map[string]*schema.NodeInstance)
	for _, item := range items {
		data[item.FlowInstanceID] = item
	}
	return data, nil
}

// QueryLastNodeInstance 查询节点实例
func (a *Flow) QueryLastNodeInstance(flowInstanceID string) (*schema.NodeInstance, error) {
	return a.FlowModel.QueryLastNodeInstance(flowInstanceID)
}
