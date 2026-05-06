package agent

type EventsManager struct {
	agentOutput   chan ResponseEvent
	agentInput    chan string
	fileChange    chan FileChangeEvent
	subagentInput chan string
	notification  chan string
}

type ChannelType int

const (
	AGENT_OUTPUT_CHANNEL ChannelType = iota
	AGENT_INPUT_CHANNEL
	FILE_DIFF_CHANNEL
	SUBAGENT_CHANNEL
	NOTIFICATION_CHANNEL
)

func (e *EventsManager) WriteToChannel(channelType ChannelType, data any) {
	switch channelType {
	case AGENT_OUTPUT_CHANNEL:
		e.agentOutput <- data.(ResponseEvent)
		return

	case AGENT_INPUT_CHANNEL:
		e.agentInput <- data.(string)
		return

	case FILE_DIFF_CHANNEL:
		e.fileChange <- data.(FileChangeEvent)
		return

	case SUBAGENT_CHANNEL:
		e.subagentInput <- data.(string)
		return

	case NOTIFICATION_CHANNEL:
		e.notification <- data.(string)
		return
	}
}

func (e *EventsManager) ReadFromChannel(channelType ChannelType) any {
	switch channelType {
	case AGENT_OUTPUT_CHANNEL:
		return <-e.agentOutput

	case AGENT_INPUT_CHANNEL:
		return <-e.agentInput

	case FILE_DIFF_CHANNEL:
		return <-e.fileChange

	case SUBAGENT_CHANNEL:
		return <-e.subagentInput

	case NOTIFICATION_CHANNEL:
		return <-e.notification
	}

	return nil
}

var EventManager = CreateEventManager()

func CreateEventManager() EventsManager {
	return EventsManager{
		agentOutput:   make(chan ResponseEvent),
		agentInput:    make(chan string),
		fileChange:    make(chan FileChangeEvent),
		subagentInput: make(chan string),
		notification:  make(chan string),
	}
}
