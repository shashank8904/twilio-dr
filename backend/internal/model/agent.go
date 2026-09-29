package model

type Agent struct {
	ID                 string   `json:"id" firestore:"-"`
	DisplayName        string   `json:"displayName" firestore:"displayName"`
	Languages          []string `json:"languages" firestore:"languages"`
	Domains            []string `json:"domains" firestore:"domains"`
	Status             string   `json:"status" firestore:"status"`
	CurrentCalls       int      `json:"currentCalls" firestore:"currentCalls"`
	MaxConcurrentCalls int      `json:"maxConcurrentCalls" firestore:"maxConcurrentCalls"`
}
