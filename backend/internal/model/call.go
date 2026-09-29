package model

type Call struct {
	CallSid         string `json:"callSid" firestore:"callSid"`
	Caller          string `json:"caller" firestore:"caller"`
	Called          string `json:"called" firestore:"called"`
	Language        string `json:"language" firestore:"language"`
	Domain          string `json:"domain" firestore:"domain"`
	Status          string `json:"status" firestore:"status"`
	AssignedAgentID string `json:"assignedAgentId" firestore:"assignedAgentId"`
	ConferenceName  string `json:"conferenceName" firestore:"conferenceName"`
}
