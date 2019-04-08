package main

import (
	"fmt"
	"github.com/arienmalec/alexa-go"
	"github.com/aws/aws-lambda-go/lambda"
	"os"
	"speechLiason/cloud_resources"
	"speechLiason/errors"
	"speechLiason/key_access"
	"speechLiason/queue_connect"
	"speechLiason/respond"
)

func main() {
	lambda.Start(DispatchIntents)
}

/*
TODO:
- handle default errors better
- figure out what happens when user says "no" in response to code confirmation
- figure out better way to give code without multiple prompts
 */

func DispatchIntents(request alexa.Request) (alexa.Response, error) {
	var s string
	var p string
	u := request.Session.User.UserID
	s = request.Session.Attributes["syncUserId"].(string)
	p = request.Session.Attributes["previousJobCursor"].(string)
	var r alexa.Response
	switch request.Body.Intent.Name {
	case "sync":
		c := request.Body.Intent.Slots["spokenCode"].Value
		t := request.Context.System.APIAccessToken
		d := request.Context.System.Device.DeviceID
		r = Sync(c, u, t, d)
		break
	case "createJob":
		j := request.Body.Intent.Slots["jobName"].Value
		if j == "" {
			j = p
		}
		r = CreateJob(j, u, s)
		break
	case "scan":
		j := request.Body.Intent.Slots["jobName"].Value
		if j == "" {
			j = p
		}
		_,_ =fmt.Fprintf(os.Stdout, "going to scan with user %s, and jobname %s", u, j)
		r = Scan(j, u, s)
		break
	case "emailJob":
		j := request.Body.Intent.Slots["jobName"].Value
		if j == "" {
			j = p
		}
		r = Deliver(j, u, s)
		break
	case "AMAZON.HelpIntent":
		r = respond.Welcome()
		break
	case "AMAZON.FallbackIntent":
		r = respond.Welcome()
		break
	default:
		r = respond.Welcome()
		break
	}
	return r, nil
}

func Sync(code string, voiceUserId string, token string, deviceId string) alexa.Response {
	key := key_access.GetKey()
	if err := queue_connect.InitConnection("reborne", key); err != nil {
		return errors.AnalyzeError(err)
	}
	defer queue_connect.CloseConnection()
	addr, err := cloud_resources.GetDeviceAddress(token, deviceId)
	if err != nil {
		return errors.AnalyzeError(err)
	}
	err = queue_connect.SyncAccounts(code, voiceUserId, addr)
	if err != nil {
		return errors.AnalyzeError(err)
	}
	return respond.Positively("sync your account", false, "", "")
}

func Scan(jobName, voiceUserId, syncUserId string) alexa.Response {
	key := key_access.GetKey()
	if err := queue_connect.InitConnection("reborne", key); err != nil {
		return errors.AnalyzeError(err)
	}
	defer queue_connect.CloseConnection()
	u, j, err := queue_connect.SendScanCommand(jobName, voiceUserId, syncUserId)
	if err != nil {
		return errors.AnalyzeError(err)
	}
	return respond.Positively("scan a page", false, u, j)
}

func CreateJob(jobName, voiceUserId, syncUserId string) alexa.Response {
	key := key_access.GetKey()
	if err := queue_connect.InitConnection("reborne", key); err != nil {
		return errors.AnalyzeError(err)
	}
	defer queue_connect.CloseConnection()
	u, j, err := queue_connect.SetCursor(jobName, voiceUserId, syncUserId)
	if err != nil {
		return errors.AnalyzeError(err)
	}
	return respond.Positively("create a job", false, u, j)
}

func Deliver(jobName, voiceUserId, syncUserId string) alexa.Response {
	key := key_access.GetKey()
	if err := queue_connect.InitConnection("reborne", key); err != nil {
		return errors.AnalyzeError(err)
	}
	defer queue_connect.CloseConnection()
	u, j, err := queue_connect.SendDeliveryCommand(jobName, voiceUserId, syncUserId, "email", "colinjlacy@gmail.com");
	if err != nil {
		return errors.AnalyzeError(err)
	}
	return respond.Positively("email job " + jobName, false, u, j)
}
