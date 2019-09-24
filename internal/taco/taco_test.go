package taco
import "testing"

func TestTaco(t *testing.T){
	// Arrange
	x := 1
	y := 2
	expected :=3

	//Act
	actual:= Add(x,y)

	//assert
	if actual != expected{
		t.Errorf("got %d, wanted %d", actual, expected )
	}
}