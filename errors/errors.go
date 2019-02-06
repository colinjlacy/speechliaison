package errors

import (
	"fmt"
	"github.com/arienmalec/alexa-go"
	"log"
	"speechLiason/respond"
)

type ContextualError struct {
	UserId  string
	JobName string
	Context string
	Log     string
	Action  string
}

type CursorNotFoundError ContextualError

func (e CursorNotFoundError) Error() string {
	return fmt.Sprintf("[ERROR] UserId %s, JobName %s, Context %s, Log %s", e.UserId, e.JobName, e.Context, e.Log)
}

type CursorExpiredError ContextualError

func (e CursorExpiredError) Error() string {
	return fmt.Sprintf("[ERROR] UserId %s, JobName %s, Context %s, Log %s", e.UserId, e.JobName, e.Context, e.Log)
}

type MissingJobNameError ContextualError

func (e MissingJobNameError) Error() string {
	return fmt.Sprintf("[ERROR] UserId %s, JobName %s, Context %s, Log %s", e.UserId, e.JobName, e.Context, e.Log)
}

type UnsupportedOperationError struct {
	ContextualError
	ErroneousOperation string
}

func (e UnsupportedOperationError) Error() string {
	return fmt.Sprintf("[ERROR] UserId %s, JobName %s, Context %s, Log %s", e.UserId, e.JobName, e.Context, e.Log)
}

type InvalidInputError struct {
	ContextualError
	ErroneousInput string
}

func (e InvalidInputError) Error() string {
	return fmt.Sprintf("[ERROR] UserId %s, JobName %s, Context %s, Log %s", e.UserId, e.JobName, e.Context, e.Log)
}

type SystemError ContextualError

func (e SystemError) Error() string {
	return fmt.Sprintf("[ERROR] UserId %s, JobName %s, Context %s, Log %s", e.UserId, e.JobName, e.Context, e.Log)
}

func AnalyzeError(e error) (r alexa.Response) {
	log.Print(e.Error())
	switch e.(type) {
	case CursorExpiredError:
		r = respond.Openly("It's been a while since your last scan, and you'll need to specify a job first.  You can create a new job, or use a previous job by telling me to use the job you have in mind, or telling me to scan a page to that job.  Just tell me which you'd like to do.", false)
		break
	case CursorNotFoundError:
		r = respond.Openly("You'll need to specify a job first.  You can create a new job, or tell me to scan a page to a job.  Just tell me which you'd like to do.", false)
		break
	case MissingJobNameError:
		r = respond.Openly("You'll need to specify a job first.  You can create a new job, or tell me to scan a page to a job.  Just tell me which you'd like to do.", false)
		break
	case UnsupportedOperationError:
		r = respond.Openly("Unfortunately, I can't deliver your job in the method you've selected yet.  Try asking me to email the job instead.", false)
	case InvalidInputError:
		r = respond.Openly("The email you've chosen for delivery isn't valid.  Try again with a valid email address", false)
	default:
		r = respond.Openly("There was a problem attempting your request; please try again later", false)
		break
	}
	return
}
