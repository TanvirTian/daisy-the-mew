package main

type fallBounce struct {
	gravity           float64
	maxFallVelocity   float64
	bounceRestitution float64

	isFalling    bool
	fallVel      float64
	bounceActive bool
	bounceVel    float64
	bounceSquash float64
	posYF        float64
}

func newFallBounce(startY int, tuning TuningConfig) fallBounce {
	return fallBounce{
		gravity:           tuning.Gravity,
		maxFallVelocity:   tuning.MaxFallVelocity,
		bounceRestitution: tuning.BounceRestitution,
		posYF:             float64(startY),
	}
}


func (f *fallBounce) start(currentY, floorY int) {
	f.cancel(currentY)
	if currentY < floorY {
		f.isFalling = true
	}
}

// cancel stops any in-progress fall or bounce 
func (f *fallBounce) cancel(currentY int) {
	f.isFalling = false
	f.fallVel = 0
	f.bounceActive = false
	f.bounceVel = 0
	f.bounceSquash = 0
	f.posYF = float64(currentY)
}

// update integrates gravity while falling 
func (f *fallBounce) update(floorY int) (newY int, active bool) {
	floor := float64(floorY)

	if f.isFalling {
		f.fallVel += f.gravity
		if f.fallVel > f.maxFallVelocity {
			f.fallVel = f.maxFallVelocity
		}
		f.posYF += f.fallVel
		if f.posYF >= floor {
			f.posYF = floor
			f.isFalling = false
			f.bounceActive = true
			f.bounceVel = -f.fallVel * f.bounceRestitution
			f.applyImpactSquash(f.fallVel)
		}
	} else if f.bounceActive {
		f.bounceVel += f.gravity
		f.posYF += f.bounceVel
		if f.posYF >= floor {
			f.posYF = floor
			if f.bounceVel > 1 {
				impact := f.bounceVel
				f.bounceVel = -impact * f.bounceRestitution
				f.applyImpactSquash(impact)
			} else {
				f.bounceActive = false
				f.bounceVel = 0
			}
		}
	}

	if f.bounceSquash > 0 {
		f.bounceSquash *= 0.85
		if f.bounceSquash < 0.02 {
			f.bounceSquash = 0
		}
	}

	return int(f.posYF), f.isFalling || f.bounceActive
}


func (f *fallBounce) applyImpactSquash(speed float64) {
	squash := speed * 0.02
	if squash > 0.4 {
		squash = 0.4
	}
	if squash < 0.06 {
		squash = 0.06
	}
	f.bounceSquash = squash
}

func (f *fallBounce) squash() float64 {
	return f.bounceSquash
}
