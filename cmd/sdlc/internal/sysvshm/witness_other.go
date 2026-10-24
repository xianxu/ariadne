//go:build !darwin && !linux

package sysvshm

func kernelWitness(int, int) (Witness, bool) { return Witness{}, false }
