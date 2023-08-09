package calico

import (
	"bytes"
	"log"
	"os"
	"os/exec"

	v1 "datainfra.io/ballastdata/api/v1/types"
)

type Calico interface {
	ReconcileCalico() error
}

type Cali struct {
	EksName string
	Cloud   v1.CloudType
}

func NewCali(
	eksName string,
	cloud v1.CloudType,
) Cali {
	return Cali{
		EksName: eksName,
		Cloud:   cloud,
	}

}

func (cali *Cali) ReconcileCalico() error {
	// deploy calico operator

	err := cali.deployCalicoOperator()
	if err != nil {
		return err
	}

	return nil
}

func (cali *Cali) deployCalicoOperator() error {

	file, err := os.Create(cali.EksName + ".yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	calico := exec.Command(
		"sh",
		"calico.sh",
		cali.EksName,
	)

	var outs bytes.Buffer
	var stderrs bytes.Buffer
	calico.Stdout = &outs
	calico.Stderr = &stderrs

	err = calico.Run()
	if err != nil {
		return err
	}

	return nil

}
