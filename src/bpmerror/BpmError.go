package bpmerror;

type BpmError struct {
    Message string
}

func New(base error, message string) (error){
    var myError BpmError;
    if base != nil && message != "" {
        myError = BpmError{Message: message + ". " + base.Error()}
    } else if message != "" {
        myError = BpmError{Message: message}
    } else {
        myError = BpmError{Message: base.Error()}
    }
    return &myError;
}

func (e *BpmError) Error() string {
    return e.Message;
}
