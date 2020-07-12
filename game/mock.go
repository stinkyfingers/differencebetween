package game

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stinkyfingers/differencebetween/api/testingsupport"
)

/*
mock S3 & db support
*/

func MockS3() s3iface.S3API {
	r := strings.NewReader(cards)
	fmt.Println(r.Len())
	return &testingsupport.S3{
		GetObjectOutput: &s3.GetObjectOutput{
			Body: ioutil.NopCloser(r),
		},
	}
}

var cards = `
	Total volume of pubic hair
	The ability to shit on command
	Patience
	Preparedness
	Subjective judgement
	A few dollars
	One night of exceptionally poor judgement
	Blatant racism
	Fucking versus being fucked
	A regretable trip to the bathroom
	A lifetime of bad decisions
	Five or six dollars
	A matter of perception
	Personal taste
	One wrinkled, flaccid penis
	A criminal record
	Nothing. They're the same
	The amount of alcohol required for each to seem normal
	I'd have sex with the first one
	Level of education
	Dick size
	Usefulness in bed
	How horny each makes me
	The size of the bruises they leave
	The sheer amount of regret
	Hairiness
	Wetness
	Diaper Rash
	Calloused hands
	Racism
	A couple of liver spots
	Country of origin
	Flavor
	Nutritional value
	Jolliness
	Optimistic outlook
	Expiration date
	Length
	Sexiness
	Likelihood of incarceration
	Hidden gay agenda
	Addictiveness
	Falsified documents
	Intelligence
	Odor
	Hidden Layers
	Scabs
	Taste
	Jewish-ness
	Ugliness
	Hair above vs. hair below
	Fondness for nuking the Middle East
	Just a matter of semantics
	Which Backstreet Boy each is most like
	Solely external
	Nothing a little wax and elbow grease can't spruce up
	Irritability
	Likelihood of being discovered to be a carcinogen
	Religiousness
	How much Asians admire each
	One is in Chinese food and one is full of Chinese food
	Frequency of STDs
	Purity
	Moistness
	Only one is on the dollar menu
	The latter can be tracked by scent alone
	Contagiousness
	The first reminds me of the second, but not the other way around
	Fuckability
	Flexibility
	Gross weight
	Value to America
	Everything
	A membership card
	An awkward, one-monthly event
	Tenacity
	Poshness
	Comfort level
	Craziness
	Likelihood of becoming president
	I've had the latter
	Level of insanity
	One is clearly more clever than the other
	The latter is much, much better in bed
	I wouldn't mind being trapped in a room with the former
	Neither is real, but people believe one
	Only one is worth bragging about
	Likelihood of contracting STDs
	IQ
	Ability to function properly
	Likelihood of being sold on late-night television
	Odds of being in a KFC advertisement
	Likelihood of being found in a storm drain
	Likelihood of being seen at an interstate rest stop
`
