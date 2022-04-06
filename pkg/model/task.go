package model

type TaskRow struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	State     string `json:"state"`
	URL       string `json:"url,omitempty"`
	LastError string `json:"lastError,omitempty"`
	Age       string `json:"age,omitempty"`
}
