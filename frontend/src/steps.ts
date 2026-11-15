let count = 0

// newStepID returns a unique id for a cancellable step that contacts a
// repository; CancelStep stops the step with that id even if it reaches Go
// before the step itself.
export function newStepID(): string {
  count++
  return `${Date.now()}-${count}`
}
