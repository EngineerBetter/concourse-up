package invocations

func CountTotalInvocations(invocations map[string][][]interface{}) int {
	total := 0
	for _, capturedArgs := range invocations {
		total += len(capturedArgs)
	}

	return total
}
