package khota_handler

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetTenantSizes(w http.ResponseWriter, req *http.Request) {
	kc, _ := getKubeClientset()

	cm, err := kc.CoreV1().ConfigMaps("kube-system").Get(context.TODO(), "tenant-sizes", metav1.GetOptions{})
	if err != nil {
		res := NewResponse(TenantSizeGetFail, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	sizeJson := cm.Data["size.json"]
	sendJsonResponse([]byte(sizeJson), http.StatusOK, &w)
}

func CreateTenantSizes(w http.ResponseWriter, req *http.Request) {

	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		res := NewResponse(ServerReqSizeExceed, req_error, err, http.StatusBadRequest)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	if err := req.Body.Close(); err != nil {
		res := NewResponse(ServerBodyCloseError, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	kc, _ := getKubeClientset()

	_, err = kc.CoreV1().ConfigMaps("kube-system").Create(context.TODO(), makeTenantSizeCm(string(body)), metav1.CreateOptions{})
	if err != nil {
		res := NewResponse(TenantSizeCreateFail, req_error, err, http.StatusInternalServerError)
		res.SetResponse(&w)
		res.LogResponse()
		return
	}

	res := NewResponse(TenantSizeCreateSuccess, success, nil, http.StatusOK)
	res.SetResponse(&w)
	res.LogResponse()
	return

}
