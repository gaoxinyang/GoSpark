package hadoop

import (
    "testing"
    "fmt"
)

func TestBasicRead(t *testing.T) {
//    fileURI := "hdfs://127.0.0.1:54310/user/hduser/testSplitRead.txt";
  fileURI := "hdfs://vision24.csail.mit.edu:54310/user/featureSUN397.csv";

  fmt.Printf("1\n");
	s := getSplitInfo(fileURI)
  fmt.Printf("2\n");
	nsplit := s.Len();
	
	fmt.Printf("This file has %d splits\n", nsplit);
	for it := s.Front(); it != nil; it=it.Next() {
	    slist := it.Value.([]string)
	    for j:=0; j<len(slist); j++ {
	        fmt.Printf(slist[j]);
	    }
	    fmt.Println();
	}
    
    scanner := getSplitScanner(fileURI, 0); // get the scanner of split 0
    
    for scanner.Scan() {
		fmt.Println(scanner.Text()) // read one line of data in split
	}
}

