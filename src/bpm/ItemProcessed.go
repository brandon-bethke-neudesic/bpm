package main;

type ItemProcessed struct {
    Bpm *BpmData
    Item *BpmDependency
    Name string
    Source string
    Local bool
}
