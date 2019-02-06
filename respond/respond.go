package respond

import "github.com/arienmalec/alexa-go"

func Welcome() alexa.Response {
	return createResponse("Welcome!", false)
}

func Positively(action string, endSession bool) alexa.Response {
	if action == "" {
		return createResponse("Okay", false)
	}
	return createResponse("Okay, I'll " + action + " for you.", endSession)
}

func Negatively(action string, endSession bool) alexa.Response {
	if action == "" {
		return createResponse("It seems there was a problem", false)
	}
	return createResponse("Unfortunately I couldn't " + action + " for you.", endSession)
}

func Openly(message string, endSession bool) alexa.Response {
	return createResponse(message, endSession)
}

func createResponse(outputText string, endSession bool) alexa.Response {
	var r alexa.Response
	var p alexa.Payload
	p.Type = "PlainText"
	p.Text = outputText
	r.Version = "1.0"
	r.Body = alexa.ResBody{OutputSpeech: &p, ShouldEndSession: endSession}
	return r
}
