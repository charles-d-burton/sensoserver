package helpers

import "log"

//AppendIfMissing ... only append a string to slice if it doesn't exist
func AppendStringIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

//RemoveStringByValue ... remove an element from a string slice by it's value
func RemoveStringByValue(slice []string, ele string) []string {
	for key, value := range slice {
		if value == ele {
			return append(slice[:key], slice[key+1:]...)
		}
	}
	return slice
}

//ReplaceStringByValue ... replace one string in a slice with another
func ReplaceStringByValue(slice []string, new string, old string) []string {
	log.Println("Slice before replace: ", slice)
	for key, value := range slice {
		if value == old {
			slice[key] = new
		}
	}
	log.Println("Slice after replace: ", slice)
	return slice
}
