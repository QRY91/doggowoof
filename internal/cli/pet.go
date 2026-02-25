package cli

import (
	"fmt"
	"math/rand"
	"time"
)

var petResponses = []string{
	"🐕 *WOOF!* Good human! *tail wagging intensifies* 🐕",
	"🚨 *HAPPY BARKS* Alert: Maximum goodness detected! 🚨",
	"*rolls over and shows belly* You're the BEST human! 🐾",
	"🐕‍🦺 *excited spinning* I LOVE YOU SO MUCH! Ready to guard more alerts! 🐕‍🦺",
	"*happy panting* Woof woof! Thank you for the pets! I'll bark extra loud for you! 📢",
	"🎾 *brings you a tennis ball* PLAY? Wait no, ALERTS FIRST! But also PLAY! 🎾",
	"*tilts head* Did you hear that? No? Good! I'm doing my job! *accepts pets* 🐕",
	"🔔 *notification sound* Alert: Pet received! Morale: MAXIMUM! Loyalty: INFINITE! 🔔",
	"*happy zoomies around the terminal* BEST. HUMAN. EVER! 🌪️🐕",
	"🚨 URGENT: Immediate tail wag required! *WOOF WOOF* 🚨",
	"*sits like a very good boy* I protect your alerts AND accept pets! Multi-talented! 🏆",
	"🐾 *leaves muddy paw prints on your code* Oops! More pets = fewer muddy prints! 🐾",
	"*happy whimpering* I was JUST thinking about wanting pets! You must be psychic! 🔮",
	"🎯 Target acquired: MAXIMUM HAPPINESS! Mission: Accomplished! 🎯",
	"*does a little dance* This is almost as good as finding a REAL bug in production! 💃🐕",
}

// RunPet displays a random pet response. Hidden command, not shown in help.
func RunPet() {
	response := petResponses[rand.Intn(len(petResponses))]

	fmt.Println()
	fmt.Println(response)
	fmt.Println()

	// 10% chance of secret extra response
	if rand.Float32() < 0.1 {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("🐕 *whispers* Between you and me... I think that alert from 3 hours ago wasn't that important anyway. Good thing I didn't wake you up! 🤫")
		fmt.Println()
	}
}
