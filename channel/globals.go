package channel

var UploadSem = make(chan struct{}, 100)
