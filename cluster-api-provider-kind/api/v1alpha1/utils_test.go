package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {

	It("should contain string", func() {
		slice1 := []string{"one", "two", "three"}
		slice2 := []string{"four", "five", "six"}

		for _, str := range slice1 {
			Expect(containsString(slice1, str)).To(BeTrue())
		}

		for _, str := range slice1 {
			Expect(containsString(slice2, str)).To(BeFalse())
		}
	})

	It("should remove string", func() {
		slice := []string{"one", "two", "three"}
		for _, str := range slice {
			Expect(containsString(removeString(slice, str), str)).To(BeFalse())
		}
	})

})
