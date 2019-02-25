package block
import (
	//"crypto/sha256" //FIXME: need to implement hashing
	"fmt"
)

func appendString(A string, B string)string{ //accepts two strings, returns single string
	A += B
	return A
}


func recursiveAdd(myArray[] string)string{
	if(len(myArray)==0){//if empty slice is passed
		return ""
	}else if (len(myArray) == 1){//end of function
		return myArray[0]
	}
	tempArray := []string{} //will be used to append to

	if(len(myArray)%2 == 1){ //if array is of odd length, makes it even
		myArray = append(myArray,myArray[len(myArray)-1] ) //append last element to self to make it even
	}
	
	for i:=0;i<len(myArray);i=i+2{
		tempArray = append(tempArray, appendString(myArray[i],myArray[i+1] ))//calls append, which returns a single string
	}
	if(false){//debugging code, used to print out element
		for i:=0;i<len(tempArray);i++{
			fmt.Println(tempArray[i])
		}
	}
	return recursiveAdd(tempArray) //recursion, returns single string
}
