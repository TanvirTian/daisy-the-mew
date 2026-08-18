package main

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)



type TuningConfig struct {
	WalkSpeed int `toml:"walk_speed"`
	RunSpeed  int `toml:"run_speed"`


	CloseDoubleClickTicks int `toml:"close_double_click_ticks"`
	HoldClickTicks        int `toml:"hold_click_ticks"`

	// How close the cursor must be before Daisy notices it
	ProximityRadius     float64 `toml:"proximity_radius"`
	ProximityHysteresis float64 `toml:"proximity_hysteresis"`
	MaxInactivityTicks  int     `toml:"max_inactivity_ticks"`

	// Sleep timing 
	InitialSleepDelayTicks int `toml:"initial_sleep_delay_ticks"`
	SleepCooldownMinTicks  int `toml:"sleep_cooldown_min_ticks"`
	SleepCooldownMaxTicks  int `toml:"sleep_cooldown_max_ticks"`

	// How many animation loops idle/sleep last before switching, min-max range.
	IdleLoopMin  int `toml:"idle_loop_min"`
	IdleLoopMax  int `toml:"idle_loop_max"`
	SleepLoopMin int `toml:"sleep_loop_min"`
	SleepLoopMax int `toml:"sleep_loop_max"`

	// Animation frame cadence
	FrameTicks int `toml:"frame_ticks"`

	// Fall physics.
	Gravity           float64 `toml:"gravity"`
	MaxFallVelocity   float64 `toml:"max_fall_velocity"`
	BounceRestitution float64 `toml:"bounce_restitution"`
}


func DefaultTuningConfig() TuningConfig {
	return TuningConfig{
		WalkSpeed: 16,
		RunSpeed:  90,

		CloseDoubleClickTicks: 30,
		HoldClickTicks:        10,

		ProximityRadius:     80.0,
		ProximityHysteresis: 12.0,
		MaxInactivityTicks:  180,

		InitialSleepDelayTicks: 15 * 60,
		SleepCooldownMinTicks:  30 * 60,
		SleepCooldownMaxTicks:  60 * 60,

		IdleLoopMin:  10,
		IdleLoopMax:  30,
		SleepLoopMin: 6,
		SleepLoopMax: 10,

		FrameTicks: 16,

		Gravity:           0.5,
		MaxFallVelocity:   22.0,
		BounceRestitution: 0.55,
	}
}


func LoadTuningConfig() TuningConfig {
	cfg := DefaultTuningConfig()

	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		f, createErr := os.Create("config.toml")
		if createErr != nil {
			log.Printf("[Config] Could not create config.toml, using defaults: %v", createErr)
			return cfg
		}
		if encErr := toml.NewEncoder(f).Encode(cfg); encErr != nil {
			log.Printf("[Config] Could not write config.toml, using defaults: %v", encErr)
		}
		f.Close()
		return cfg
	}

	if _, err := toml.DecodeFile("config.toml", &cfg); err != nil {
		log.Printf("[Config] Could not parse config.toml, using defaults: %v", err)
		return DefaultTuningConfig()
	}

	return cfg
}
