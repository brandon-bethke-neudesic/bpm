package main;
type SubCommand interface {
    Execute() error
}
