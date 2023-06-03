package model

// QuestStatus enumerates a player's status of a quest.
type QuestStatus int

const (
	QuestStatusNotStarted QuestStatus = iota
	QuestStatusInProgress
	QuestStatusFinished
)
