package main

import "testing"

// Every effect must spawn something and fully clean up afterwards.
func TestEffectsSpawnAndExpire(t *testing.T) {
	for i := range effects {
		var s System
		s.Play(&effects[i], 400, 300)
		s.Update()
		if s.Len() == 0 {
			t.Fatalf("%s spawned nothing", effects[i].Name)
		}
		for range 600 {
			s.Update()
		}
		if s.Len() != 0 {
			t.Fatalf("%s left %d things alive", effects[i].Name, s.Len())
		}
	}
}

// Zero-value Options must not panic (empty Kinds/Colors, LifeMax < LifeMin).
func TestZeroOptions(t *testing.T) {
	var s System
	s.Emit(Options{Count: 3, LifeMin: 5}, 0, 0)
	for range 10 {
		s.Update()
	}
	if s.Len() != 0 {
		t.Fatal("particles did not expire")
	}
}

func TestBallRadiusStartsAndEndsAtZero(t *testing.T) {
	b := effects[0].Ball
	if r := ballRadius(b, 0); r != 0 {
		t.Fatalf("start radius %v, want 0", r)
	}
	if r := ballRadius(b, b.Ticks); r > 0.001 {
		t.Fatalf("end radius %v, want 0", r)
	}
}

// addFan must only reference vertices it added.
func TestAddFanIndices(t *testing.T) {
	var s System
	s.addFan(0, 0, 0, 1, 1, unitStar, party[0])
	s.addFan(5, 5, 1, 2, 2, unitTri, party[1])
	if len(s.is) != 3*(len(unitStar)+len(unitTri)) {
		t.Fatalf("got %d indices", len(s.is))
	}
	for _, i := range s.is {
		if int(i) >= len(s.vs) {
			t.Fatalf("index %d out of range %d", i, len(s.vs))
		}
	}
}
