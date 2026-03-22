package agent

type EventsManager struct {
	agentOutput chan ResponseEvent
	agentInput  chan string
	fileChange  chan FileChangeEvent
}

type ChannelType int

const (
	AGENT_OUTPUT_CHANNEL ChannelType = iota
	AGENT_INPUT_CHANNEL
	FILE_DIFF_CHANNEL
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
	}

	return nil
}

var EventManager = CreateEventManager()

func CreateEventManager() EventsManager {
	return EventsManager{
		agentOutput: make(chan ResponseEvent),
		agentInput:  make(chan string),
		fileChange:  make(chan FileChangeEvent),
	}
}
