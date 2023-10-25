package khota_handler

import (
	"encoding/json"
	"net/http"

	klog "k8s.io/klog/v2"
)

type Response struct {
	Msg        CustomMsg
	Status     string
	StatusCode int
	Err        error
}

func NewResponse(msg CustomMsg, status string, err error, statusCode int) *Response {
	return &Response{Msg: msg, Status: status, StatusCode: statusCode, Err: err}
}
func (res *Response) LogResponse() {
	if res.Err != nil {
		klog.Errorf("ErrMsg: [%s], Status: [%s], Error: [%s], statusCode [%d]", res.Msg, res.Status, res.Err.Error(), res.StatusCode)
	} else {
		klog.Infof("Msg: [%s], Status: [%s], Error: [%s], statusCode [%d]", res.Msg, res.Status, "", res.StatusCode)
	}

}
func (res *Response) SetResponse(w *http.ResponseWriter) {
	(*w).Header().Set("Content-Type", "application/json; charset=UTF-8")
	(*w).WriteHeader(res.StatusCode)
	json.NewEncoder(*w).Encode(res)
}
func (res *Response) SetMsgResponse(w *http.ResponseWriter) {
	jsonData, err := json.Marshal(res.Msg)
	if err != nil {
		klog.Errorf("ErrMsg: Unable to parse response Message json [%s], Error: [%s]", res.Msg, err.Error())
		res.StatusCode = http.StatusInternalServerError
	}
	(*w).Header().Set("Content-Type", "application/json; charset=UTF-8")
	(*w).WriteHeader(res.StatusCode)
	(*w).Write(jsonData)
}
