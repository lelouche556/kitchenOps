package models

type TaskStatus string

const (
	TaskUnassigned TaskStatus = "UNASSIGNED"
	TaskAssigned   TaskStatus = "ASSIGNED"
	TaskStarted    TaskStatus = "STARTED"
	TaskCompleted  TaskStatus = "COMPLETED"
)

type OrderStatus string

const (
	OrderConfirmed     OrderStatus = "CONFIRMED"
	OrderInProgress    OrderStatus = "IN_PROGRESS"
	OrderPartCompleted OrderStatus = "PART_COMPLETED"
	OrderCompleted     OrderStatus = "COMPLETED"
)

type EventType string

const (
	EventTaskCreated   EventType = "TASK_CREATED"
	EventTaskReady     EventType = "TASK_READY"
	EventTaskAssigned  EventType = "TASK_ASSIGNED"
	EventTaskStarted   EventType = "TASK_STARTED"
	EventTaskCompleted EventType = "TASK_COMPLETED"
	EventTaskRequeued  EventType = "TASK_REQUEUED"
)

type DomainEvent struct {
	Type      EventType
	Timestamp int
	Payload   string
}
