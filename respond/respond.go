package respond

import "github.com/arienmalec/alexa-go"

func Welcome() alexa.Response {
	return createResponse("Welcome!", false, "", "")
}

func Positively(action string, endSession bool, syncUserId string, jobName string) alexa.Response {
	if action == "" {
		return createResponse("Okay", false, syncUserId, jobName)
	}
	return createResponse("Okay, I'll " + action + " for you.", endSession, syncUserId, jobName)
}

func Negatively(action string, endSession bool, syncUserId string, jobName string) alexa.Response {
	if action == "" {
		return createResponse("It seems there was a problem", false, syncUserId, jobName)
	}
	return createResponse("Unfortunately I couldn't " + action + " for you.", endSession, syncUserId, jobName)
}

func Openly(message string, endSession bool, syncUserId string, jobName string) alexa.Response {
	return createResponse(message, endSession, syncUserId, jobName)
}

func createResponse(outputText string, endSession bool, syncUserId string, jobName string) alexa.Response {
	var r alexa.Response
	var p alexa.Payload
	p.Type = "PlainText"
	p.Text = outputText
	r.Version = "1.0"
	r.Body = alexa.ResBody{OutputSpeech: &p, ShouldEndSession: endSession}
	if !endSession {
		r.SessionAttributes["syncUserId"] = syncUserId
		r.SessionAttributes["previousJobCursor"] = jobName
	}
	return r
}
