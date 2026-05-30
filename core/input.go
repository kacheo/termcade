package core

type InputHandler struct {
	keys    []string
	timeout int
}

func NewInputHandler() *InputHandler {
	return &InputHandler{
		keys:    []string{},
		timeout: 0,
	}
}

func (h *InputHandler) HandleKey(key string) {
	h.keys = append(h.keys, key)
}

func (h *InputHandler) GetKey() (string, bool) {
	if len(h.keys) == 0 {
		return "", false
	}
	k := h.keys[0]
	h.keys = h.keys[1:]
	return k, true
}

func (h *InputHandler) Clear() {
	h.keys = []string{}
}