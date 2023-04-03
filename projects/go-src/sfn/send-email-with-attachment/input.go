package main

type Input struct {
	Parameter map[string]string `json:"parameter"`
	Template  string            `json:"template"`
	Sender    string            `json:"sender"`
	Recipient string            `json:"recipient"`
	Documents []Document        `json:"documents"`
	Raw       RenderedTemplate  `json:"raw"`
}

type Document struct {
	Bucket   string `json:"bucket"`
	Key      string `json:"key"`
	FileName string `json:"fileName"`
}

type RenderedTemplate struct {
	RenderedTemplate string `json:"RenderedTemplate"`
}
