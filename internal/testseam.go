package gopher

// SetFrameSize rescales every layout constant calibrated against the
// 1024-pixel production sprite. Intended for tests only — call from a
// TestMain in the consuming package to shrink integration-test sources from
// 1024×1024 to e.g. 64×64, which speeds up PNG encode by ~30×.
//
// Concurrency contract: call exactly once from TestMain before m.Run, and
// never from a t.Parallel test body. SetFrameSize mutates eight package
// globals (frameSize, footBand{Top,Bot}, armBand{Top,Bot}, bobAmp, footLift,
// armLift) that the sprite generators read without locking.
func SetFrameSize(size int) {
	scale := float64(size) / 1024.0
	scaleInt := func(v int) int {
		out := int(float64(v) * scale)
		if out < 1 {
			return 1
		}
		return out
	}
	frameSize = size
	footBandTop = scaleInt(820)
	footBandBot = size
	armBandTop = scaleInt(450)
	armBandBot = scaleInt(750)
	bobAmp = scaleInt(50)
	footLift = scaleInt(150)
	armLift = scaleInt(120)
}
