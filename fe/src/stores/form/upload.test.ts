import { getMaxFileSize } from "./upload"

const testGetMaxFileSize = () => {
  // assert it doesn't return 0 or throw an error
  expect(getMaxFileSize()).not.toBe(0)

  // assert it returns a number
  expect(typeof getMaxFileSize()).toBe("number")
}
