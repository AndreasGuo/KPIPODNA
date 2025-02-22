package test

import (
	"GoDNA/algorithm"
	DNAType "GoDNA/algorithm/dnatype"
	"log"
	"math"
	"math/rand"
	"testing"
)

func TestOraginForaging(T *testing.T) {
	fitChan := DNAType.CreateWorker(100, 100, 10)
	defer fitChan.Close()

	// init dna set
	var dnaSet = baseDNAs()
	var zmin = make([]float64, 5)
	for j := range zmin {
		zmin[j] = 400
	}

	for i := range dnaSet {
		fitFunc := DNAType.FitnessCall(dnaSet, i, fitChan, 2e-2, true)
		singleInvSlice := []*DNAType.DNAAgent{dnaSet[i]}
		fits, _ := fitFunc(singleInvSlice)
		dnaSet[i].SetObjs(fits[0])
	}

	index := 2
	//fmt.Println("To Optimize: ", index)
	fitFunc := DNAType.FitnessCall(dnaSet, index, fitChan, 2e-2, true)
	alg := PO{Pop: nil, MaxIteration: 200}
	pop := new(DNAType.DNAPopulation)
	pop.SetConfig(200, 20, 4, 0, 3)
	pop.SetFitFunc(fitFunc)
	alg.Initialize(pop, dnaSet[index])
	inv := alg.IterationForaging(false, true, false)
	PrintDNASet(dnaSet, fitChan)
	dnaSet[index] = inv
}

func TestAdjustForaging(T *testing.T) {
	fitChan := DNAType.CreateWorker(100, 100, 10)
	defer fitChan.Close()

	// init dna set
	var dnaSet = baseDNAs()
	var zmin = make([]float64, 5)
	for j := range zmin {
		zmin[j] = 400
	}

	for i := range dnaSet {
		fitFunc := DNAType.FitnessCall(dnaSet, i, fitChan, 2e-2, true)
		singleInvSlice := []*DNAType.DNAAgent{dnaSet[i]}
		fits, _ := fitFunc(singleInvSlice)
		dnaSet[i].SetObjs(fits[0])
	}

	index := 2
	//fmt.Println("To Optimize: ", index)
	fitFunc := DNAType.FitnessCall(dnaSet, index, fitChan, 2e-2, true)
	alg := PO{Pop: nil, MaxIteration: 200}
	pop := new(DNAType.DNAPopulation)
	pop.SetConfig(200, 20, 4, 0, 3)
	pop.SetFitFunc(fitFunc)
	alg.Initialize(pop, dnaSet[index])
	inv := alg.IterationForaging(false, false, false)
	PrintDNASet(dnaSet, fitChan)
	dnaSet[index] = inv
}

// PO + NDSort + Knee Point
func (po *PO) IterationForaging(hyperPlaneNorm bool, origin bool, cd bool) *DNAType.DNAAgent {
	logger := log.Default()
	islog := false
	fits := po.Pop.Fit()
	ZMin := po.Pop.ZMin()
	selectedIndex, _ := algorithm.NDKPSort(fits, ZMin, po.Pop.Size(), hyperPlaneNorm, cd)

	gbest := po.Pop.At(selectedIndex).Variance()
	writeCSV(fits, "foraging", -1, origin, selectedIndex)
	for it := 0; it < int(po.MaxIteration); it++ {
		oldPop := po.Pop.Clone()
		popMean := Mean(po.Pop)
		for i := 0; i < int(po.Pop.Size()); i++ {
			if i == selectedIndex {
				continue
			}
			st := rand.Intn(4)
			x := po.Pop.At(i).Variance()
			if islog {
				logger.Println("it: ", it, "st: ", st)
				logger.Println("origin variance: ", x)
			}

			if origin {
				ost0(x, gbest, popMean, po.Pop.VarianceDim(), it, po.MaxIteration)
			} else {
				st0(x, gbest, popMean, po.Pop.VarianceDim(), it, po.MaxIteration)
			}

			out := make([]int, len(x))
			for k := range x {
				out[k] = int(math.Round(x[k]))
			}
			if islog {
				logger.Println("After st: ", out)
			}
			//boundary control
			for j := 0; j < po.Pop.VarianceDim(); j++ {
				x[j] = max(x[j], po.Pop.LB())
				x[j] = min(x[j], po.Pop.UB())
			}
			for k := range x {
				out[k] = int(math.Round(x[k]))
			}
			if islog {
				logger.Println("After boundary: ", out)
			}
			po.Pop.UpdatePosition(i, x)
		}
		po.Pop.PostWork()

		po.Pop.Join(oldPop)

		fits = po.Pop.Fit()
		ZMin = po.Pop.ZMin()
		bestIdx, selectedIndex := algorithm.NDKPSort(fits, ZMin, po.Pop.Size()/2, hyperPlaneNorm, cd)
		gbest = po.Pop.At(bestIdx).Variance()
		po.Pop.Select(selectedIndex)
		if it%50 == 0 || it == po.MaxIteration-1 {
			csvFits := [][]float64{}
			for _, idx := range selectedIndex {
				csvFits = append(csvFits, fits[idx])
			}
			writeCSV(csvFits, "foraging", it, origin, bestIdx)
		}
	}

	return po.Pop.At(selectedIndex)
}
