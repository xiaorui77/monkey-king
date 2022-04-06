package model

type TaskRow struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	State     string `json:"state"`
	URL       string `json:"url"`
	LastError string `json:"lastError,omitempty"`
	Age       string `json:"age"`
}
