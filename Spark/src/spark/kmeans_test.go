package spark

import (
  "testing"
  "fmt"
  "math/rand"
  "strings"
  "strconv"
  "encoding/gob"
)

// Format [PicIndex],[feature1],[feature2],[feature3],[feature4],[feature5]...
func (f *UserFunc) MapLineToFloatVectorCSV(line interface{}) interface{} {
  fieldTexts := strings.FieldsFunc(line.(KeyValue).Value.(string), func(c rune) bool { return c == ',' })
  
  vecs := make(Vector, len(fieldTexts)-1)
  for i := range vecs {
    vecs[i], _ = strconv.ParseFloat(fieldTexts[i+1], 64)
  }
  return vecs
}

// Format [feature1] [feature2] [feature3] [feature4] [feature5]...
func (f *UserFunc) MapLineToFloatVector(line interface{}) interface{} {
  fieldTexts := strings.Fields(line.(KeyValue).Value.(string))
  
  vecs := make(Vector, len(fieldTexts))
  for i := range vecs {
    vecs[i], _ = strconv.ParseFloat(fieldTexts[i], 64)
  }
  return vecs
}

// Format [PicIndex],[CategoryIndex],[feature1],[feature2],[feature3],[feature4],[feature5]...
func (f *UserFunc) MapLineToFloatVectorWithCat(line interface{}) interface{} {
  fieldTexts := strings.Fields(line.(KeyValue).Value.(string))
  
  vecs := make(Vector, len(fieldTexts))
  for i := range vecs {
    vecs[i], _ = strconv.ParseFloat(fieldTexts[i+2], 64)
  }
  return vecs
}

// Format [PicIndex],[CategoryIndex],[feature1],[feature2],[feature3],[feature4],[feature5]...
func (f *UserFunc) MapLineToCat(line interface{}) interface{} {
  fieldTexts := strings.Fields(line.(KeyValue).Value.(string))
  
  return fieldTexts[1]
}

func (f *UserFunc) MapToClosestCenter(line interface{}, userData interface{}) interface{} {
  p := line.(KeyValue).Value.(Vector)
  centers := userData.(KeyValue).Value.([]Vector)
  
  minDist := p.EulaDistance(centers[0])
  minIndex := 0
  for i := 1; i < len(centers); i++ {
    dist := p.EulaDistance(centers[i])
    if dist < minDist {
      minDist = dist
      minIndex = i
    }
  }
  return KeyValue{
                Key:   minIndex,
                Value: CenterCounter{p, 1},
         }
}

type CenterCounter struct {
    X     Vector
    Count int
}

func (f *UserFunc) AddCenterWCounter(x interface{}, y interface{}) interface{} {
  cc1 := x.(KeyValue).Value.(CenterCounter)
  cc2 := y.(KeyValue).Value.(CenterCounter)
  return CenterCounter{
    X:     cc1.X.Plus(cc2.X),
    Count: cc1.Count + cc2.Count,
  }
}

func (f *UserFunc) AvgCenter(x interface{}) interface{} {
  keyValue := x.(KeyValue).Value.(KeyValue)
  cc := keyValue.Value.(CenterCounter)
  return KeyValue{
    Key:   keyValue.Key,
    Value: cc.X.Divide(float64(cc.Count)),
  }
}


func TestKMeans(t *testing.T) {
  c := NewContext("kmeans")
  defer c.Stop()
  
  gob.Register(CenterCounter{})
  gob.Register([]Vector{})
  
  D := 4096
  K := 397
  //MIN_DIST := 0.01

  centers := make([]Vector, K)
  for i := range centers {
    center := make(Vector, D)
    for j := range center {
      center[j] = rand.Float64()
    }
    centers[i] = center
  }
  fmt.Println(centers)
  
  pointsText := c.TextFile("hdfs://vision24.csail.mit.edu:54310/user/featureSUN397_combine.csv")
  //pointsText := c.TextFile("hdfs://localhost:54310/user/kmean_data.txt"); pointsText.name = "pointsText"
  
  points := pointsText.Map("MapLineToFloatVectorCSVWithCat").Cache();  points.name = "points"
  
  // run one kmeans iteration
  // points (x,y) -> (index of the closest center, )
  
  for i := 0; i < 10; i++ {
    fmt.Println("Iter:", i)
	  mappedPoints := points.MapWithData("MapToClosestCenter", centers); mappedPoints.name = "mappedPoints"   
	  sumCenters := mappedPoints.ReduceByKey("AddCenterWCounter") ; sumCenters.name = "sumCenters"  
		newCenters := sumCenters.Map("AvgCenter")                   ; newCenters.name = "newCenters"  
		newCentersCollected := newCenters.Collect()                 
		for i:=0; i<len(newCentersCollected); i++ {
		  centers[i] = *(newCentersCollected[i].(KeyValue).Value.(*Vector))
		}
    fmt.Printf("Round %v Centers: %v\n", i, centers)
  }
  KmeansLabels := mappedPoints.Collect()
  
  trueLabels := pointsText.Map("MapLineToCat");   trueLabels.name = "trueLabels"
  TrueLabels := trueLabels.Collect()
  
  fout, err := os.Create("KmeansOutput-CompareLabels.txt")
  defer fout.Close()
  for i := 0; i < len(KmeansLabels); i++ {
    fout.WriteString( fmt.Sprintf("%d %s\n", KmeansLabels[i].(KeyValue).Key.(int), TrueLabels[i]) ) 
  }
}













/////////////////////////// deprecated tests


func xTestKMeansStepByStep(t *testing.T) {
  c := NewContext("kmeans")
  defer c.Stop()
  
  gob.Register(CenterCounter{})
  gob.Register([]Vector{})
  
  D := 4
  K := 3

  centers := make([]Vector, K)
  for i := range centers {
    center := make(Vector, D)
    for j := range center {
      center[j] = rand.Float64()
    }
    centers[i] = center
  }
  fmt.Println(centers)
  
  //pointsText := c.TextFile("hdfs://vision24.csail.mit.edu:54310/user/featureSUN397.csv")
  pointsText := c.TextFile("hdfs://localhost:54310/user/kmean_data.txt"); pointsText.name = "pointsText"
  //pointsText := c.TextFile("hdfs://localhost:54310/user/hduser/testSplitRead.txt"); pointsText.name = "pointsText"
  
  points := pointsText.Map("MapLineToFloatVector").Cache();  points.name = "points"
  
  // run one kmeans iteration
  // points (x,y) -> (index of the closest center, )
  mappedPoints := points.MapWithData("MapToClosestCenter", centers); mappedPoints.name = "mappedPoints"   
  sumCenters := mappedPoints.ReduceByKey("AddCenterWCounter") ; sumCenters.name = "sumCenters"  
  
  collect := mappedPoints.Collect()
  
  fmt.Println("Collected:", collect)
  
}


func xTestKMeansOneIter(t *testing.T) {
  c := NewContext("kmeans")
  defer c.Stop()
  
  gob.Register(CenterCounter{})
  gob.Register([]Vector{})
  
  D := 4
  K := 3
  //MIN_DIST := 0.01

  centers := make([]Vector, K)
  for i := range centers {
    center := make(Vector, D)
    for j := range center {
      center[j] = rand.Float64()
    }
    centers[i] = center
  }
  fmt.Println(centers)
  
  //pointsText := c.TextFile("hdfs://vision24.csail.mit.edu:54310/user/featureSUN397.csv")
  pointsText := c.TextFile("hdfs://localhost:54310/user/kmean_data.txt"); pointsText.name = "pointsText"
  //pointsText := c.TextFile("hdfs://localhost:54310/user/hduser/testSplitRead.txt"); pointsText.name = "pointsText"
  
  points := pointsText.Map("MapLineToFloatVector").Cache();  points.name = "points"
  
  // run one kmeans iteration
  // points (x,y) -> (index of the closest center, )
  mappedPoints := points.MapWithData("MapToClosestCenter", centers); mappedPoints.name = "mappedPoints"   
  sumCenters := mappedPoints.ReduceByKey("AddCenterWCounter") ; sumCenters.name = "sumCenters"  
  newCenters := sumCenters.Map("AvgCenter")                   ; newCenters.name = "newCenters"  
  newCentersCollected := newCenters.Collect()                 
  for i:=0; i<len(newCentersCollected); i++ {
    centers[i] = *(newCentersCollected[i].(KeyValue).Value.(*Vector))
  }
  
  fmt.Println("Final Centers:", centers)
}
