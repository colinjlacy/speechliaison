package errors

import (
	"fmt"
	"github.com/arienmalec/alexa-go"
	"log"
	"speechLiason/respond"
)

type ContextualError struct {
	UserId     string
	JobName    string
	Context    string
	Log        string
	Action     string
	SpokenCode string
}

type CursorNotFoundError ContextualError

func (e CursorNotFoundError) Error() string {
	return fmt.Sprintf("[ERROR] UserId %s, JobName %s, Context %s, Log %s", e.UserId, e.JobName, e.Context, e.Log)
}

type SyncDocNotFoundError ContextualError

func (e SyncDocNotFoundError) Error() string {
	return fmt.Sprintf("[ERROR] SpokenCode %s, Context %s, Log %s", e.SpokenCode, e.Context, e.Log)
}

type CursorExpiredError ContextualError

func (e CursorExpiredError) Error() string {
	return fmt.Sprintf("[ERROR] UserId %s, JobName %s, Context %s, Log %s", e.UserId, e.JobName, e.Context, e.Log)
}

type SyncDocExpiredError ContextualError

func (e SyncDocExpiredError) Error() string {
	return fmt.Sprintf("[ERROR] SpokenCode %s, Context %s, Log %s", e.SpokenCode, e.Context, e.Log)
}

type UserAccountNotSyncedError ContextualError

func (e UserAccountNotSyncedError) Error() string {
	return fmt.Sprintf("[ERROR] UserId %s, Context %s, Log %s", e.UserId, e.Context, e.Log)
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

func AnalyzeError(e error, syncUserId, jobName string) (r alexa.Response) {
	log.Print(e.Error())
	switch e.(type) {
	case CursorExpiredError:
		r = respond.Openly("It's been a while since your last scan, and you'll need to specify a job first.  You can create a new job, or use a previous job by telling me to use the job you have in mind, or telling me to scan a page to that job.  Just tell me which you'd like to do.", false, syncUserId, jobName)
		break
	case CursorNotFoundError:
		r = respond.Openly("You'll need to specify a job first.  You can create a new job, or tell me to scan a page to an existing job.  Just tell me which you'd like to do.", false, syncUserId, jobName)
		break
	case SyncDocNotFoundError:
		r = respond.Openly("Unfortunately I was not able to find a match to the code you specified.  Please speak the prompt given on your screen again, or click cancel, and retry the sync process.", false, syncUserId, jobName)
		break
	case SyncDocExpiredError:
		r = respond.Openly("While I was able to find a matching code to the one you spoke, it has unfortunately expired.  Please click cancel, and retry the sync process while making sure to speak the code given within a few minutes.", false, syncUserId, jobName)
		break
	case UserAccountNotSyncedError:
		r = respond.Openly("You have not yet synced your Echo device to your Reborne account.  Please open the Reborne dashboard in your browser, and start the sync process by clicking on the user icon on the screen.", false, "", jobName)
		break
	case MissingJobNameError:
		r = respond.Openly("You'll need to specify a job first.  You can create a new job, or tell me to scan a page to a job.  Just tell me which you'd like to do.", false, syncUserId, jobName)
		break
	case UnsupportedOperationError:
		r = respond.Openly("Unfortunately, I can't deliver your job in the method you've selected yet.  Try asking me to email the job instead.", false, syncUserId, jobName)
		break
	case InvalidInputError:
		r = respond.Openly("The email you've chosen for delivery isn't valid.  Try again with a valid email address", false, syncUserId, jobName)
		break
	default:
		r = respond.Openly("There was a problem attempting your request; please try again later", false, syncUserId, jobName)
		break
	}
	return
}
