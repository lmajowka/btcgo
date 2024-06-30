package utils

import "bytes"

func Contains(slice [][]byte, item []byte) bool {
    for _, a := range slice {
        if bytes.Equal(a, item) {
            return true
        }
    }
    return false
}
