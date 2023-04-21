package temporaltui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	proto "github.com/gogo/protobuf/proto"

	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	temporalClient "go.temporal.io/sdk/client"
	"go.temporal.io/server/common/codec"

	"github.com/neomantra/tempted/internal/tui/components/page"
	"github.com/neomantra/tempted/internal/tui/formatter"
	"github.com/neomantra/tempted/internal/tui/message"
)

type WorkflowKey struct {
	WorkflowID        string
	RunID             string
	ParentNamespaceID string
}

///////////////////////////////////////////////////////////////////////////////

func FetchWorkflowExecutions(client temporalClient.Client) tea.Cmd {
	return func() tea.Msg {
		// fetch all the workflows
		query := "" // TODO:
		workflowExecutions, err := getWorkflowExecutions(context.Background(), client, query)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		// sort.Slice(workflowsResponses, func(x, y int) bool {
		// 	wfX := workflowsResponses[x]
		// 	wfY := workflowsResponses[y]
		// 	// if wfX..Name == secondJob.Name {
		// 	// 	return firstJob.Namespace < secondJob.Namespace
		// 	// }
		// 	// return jobResults[x].Name < jobResults[y].Name
		// })

		tableHeader, allPageData := workflowExecutionsAsTable(workflowExecutions)
		return PageLoadedMsg{
			Page:        WorkflowsPage,
			TableHeader: tableHeader,
			AllPageRows: allPageData}
	}
}

func FetchWorkflowDetails(key WorkflowKey, client temporalClient.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		resp, err := client.DescribeWorkflowExecution(ctx, key.WorkflowID, key.RunID)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		descBytes, err := json.Marshal(resp)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		pretty := formatter.PrettyJsonStringAsLines(string(descBytes))

		var workflowDetailsPageData []page.Row
		for _, row := range pretty {
			workflowDetailsPageData = append(workflowDetailsPageData, page.Row{Key: "", Row: row})
		}

		return PageLoadedMsg{
			Page:        WorkflowDetailsPage,
			TableHeader: []string{},
			AllPageRows: workflowDetailsPageData,
		}
	}
}

// info := resp.GetWorkflowExecutionInfo()
// executionInfo := &clispb.WorkflowExecutionInfo{
// 	Execution:            info.GetExecution(),
// 	Type:                 info.GetType(),
// 	CloseTime:            info.GetCloseTime(),
// 	StartTime:            info.GetStartTime(),
// 	Status:               info.GetStatus(),
// 	HistoryLength:        info.GetHistoryLength(),
// 	ParentNamespaceId:    info.GetParentNamespaceId(),
// 	ParentExecution:      info.GetParentExecution(),
// 	Memo:                 info.GetMemo(),
// 	SearchAttributes:     convertSearchAttributes(c, info.GetSearchAttributes()),
// 	AutoResetPoints:      info.GetAutoResetPoints(),
// 	StateTransitionCount: info.GetStateTransitionCount(),
// 	ExecutionTime:        info.GetExecutionTime(),
// }

// var pendingActivitiesStr []*clispb.PendingActivityInfo
// for _, pendingActivity := range resp.GetPendingActivities() {
// 	pendingActivityStr := &clispb.PendingActivityInfo{
// 		ActivityId:         pendingActivity.GetActivityId(),
// 		ActivityType:       pendingActivity.GetActivityType(),
// 		State:              pendingActivity.GetState(),
// 		ScheduledTime:      pendingActivity.GetScheduledTime(),
// 		LastStartedTime:    pendingActivity.GetLastStartedTime(),
// 		LastHeartbeatTime:  pendingActivity.GetLastHeartbeatTime(),
// 		Attempt:            pendingActivity.GetAttempt(),
// 		MaximumAttempts:    pendingActivity.GetMaximumAttempts(),
// 		ExpirationTime:     pendingActivity.GetExpirationTime(),
// 		LastFailure:        convertFailure(pendingActivity.GetLastFailure()),
// 		LastWorkerIdentity: pendingActivity.GetLastWorkerIdentity(),
// 	}

// 	if pendingActivity.GetHeartbeatDetails() != nil {
// 		pendingActivityStr.HeartbeatDetails = stringify.AnyToString(pendingActivity.GetHeartbeatDetails(), true, 0, customDataConverter())
// 	}
// 	pendingActivitiesStr = append(pendingActivitiesStr, pendingActivityStr)
// }

// return &clispb.DescribeWorkflowExecutionResponse{
// 	ExecutionConfig:       resp.ExecutionConfig,
// 	WorkflowExecutionInfo: executionInfo,
// 	PendingActivities:     pendingActivitiesStr,
// 	PendingChildren:       resp.PendingChildren,
// 	PendingWorkflowTask:   resp.PendingWorkflowTask,
// }

func prettyPrintJSONObject(o interface{}) (string, error) {
	var b []byte
	var err error
	if pb, ok := o.(proto.Message); ok {
		encoder := codec.NewJSONPBIndentEncoder("  ")
		b, err = encoder.Encode(pb)
	} else {
		b, err = json.MarshalIndent(o, "", "  ")
	}

	if err != nil {
		return "", err
	}
	return string(b), nil
}

///////////////////////////////////////////////////////////////////////////////

// getWorkflows calls ListWorkflow with query and gets all workflow exection infos in a list.
func getWorkflowExecutions(ctx context.Context, c temporalClient.Client, query string) ([]*workflowpb.WorkflowExecutionInfo, error) {
	var nextPageToken []byte
	var workflowExecutions []*workflowpb.WorkflowExecutionInfo
	for {
		resp, err := c.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Query:         query,
			NextPageToken: nextPageToken,
		})
		if err != nil {
			return nil, err
		}
		workflowExecutions = append(workflowExecutions, resp.Executions...)
		nextPageToken = resp.NextPageToken
		if len(nextPageToken) == 0 {
			return workflowExecutions, nil
		}
	}
}

func workflowExecutionsAsTable(infos []*workflowpb.WorkflowExecutionInfo) ([]string, []page.Row) {
	var workflowExecutionRows [][]string
	var keys []string
	for _, info := range infos {
		workflowExecutionRows = append(workflowExecutionRows, []string{
			info.Type.Name,
			info.Execution.WorkflowId,
			info.Execution.RunId,
			info.TaskQueue,
			formatter.FormatTimePtr(info.StartTime),
			formatter.FormatTimePtr(info.ExecutionTime),
			formatter.FormatTimePtr(info.CloseTime),
		})
		keys = append(keys, formatWorkflowKey(info))
	}

	columns := []string{"Type", "Workflow ID", "Run ID", "Task Queue", "Start Time", "Exec Time", "End Time"}
	table := formatter.GetRenderedTableAsString(columns, workflowExecutionRows)

	var rows []page.Row
	for idx, row := range table.ContentRows {
		rows = append(rows, page.Row{Key: keys[idx], Row: row})
	}

	return table.HeaderRows, rows
}

///////////////////////////////////////////////////////////////////////////////

func formatWorkflowKey(info *workflowpb.WorkflowExecutionInfo) string {
	return fmt.Sprintf("%s %s %s", info.Execution.WorkflowId, info.Execution.RunId, info.ParentNamespaceId)
}

func WorkflowKeyFromString(key string) WorkflowKey {
	split := strings.Split(key, " ")
	return WorkflowKey{split[0], split[1], split[2]}
}
