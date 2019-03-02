package keys

import (
	"encoding/json"
	"math/big"
	"io/ioutil"
	"os"
)

/*=================================================================================================
* Purpose: Reads public and private keys from a json file, converts to big.Int                    *
* Parameters: filename, a string containing the filename for the json file. usually               *
*  ./.keys.dat.json                                                                               *
* Returns: private key, a big.Int and public key, a big.Int                                       *
=================================================================================================*/
func GetKeys(filename string) (*big.Int, *big.Int) {
	// Reads json file into b_string, if any errors occur, abort
	if b_string, err := ioutil.ReadFile(filename); err == nil {
		// Create message variable to hold json data
		type keyStruct struct {
			Private string
			Public string
		}
		var keys keyStruct
		// Load json data from text file
		err = json.Unmarshal(b_string, &keys)

		// Sets priv_key to a new big int, and converts sting from json file into big int
		private_key := new(big.Int)
		private_key.SetString(keys.Private, 16)

		// Sets pub_key to a new big int, and converts sting from json file into big int
		public_key := new(big.Int)
		public_key.SetString(keys.Public, 16)

		// Returns pub_key and priv_key
		return private_key, public_key
	} else {
		// Fatal exit, placeholder
		os.Exit(-1)
		return big.NewInt(-1), big.NewInt(-1)
	}
}