package slack

type Payload struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Text   string `json:"text"`
	Color  string `json:"color"`
	Footer string `json:"footer,omitempty"`
}

func (p *Payload) Attach(attachments []Attachment) *Payload {
	p.Attachments = append(p.Attachments, attachments...)
	return p
}
