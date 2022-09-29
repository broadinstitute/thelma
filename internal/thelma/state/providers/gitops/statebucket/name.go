package statebucket

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"math/rand"
	"strings"
)

const maxUniqueNameAttempts = 15
const defaultPrefix = "bee"
const separator = "-"

var adjectives = []string{"able", "above", "absurd", "ace", "active", "actual", "alert", "alive", "aloof", "amazed", "ample", "amused", "apt", "awake", "aware", "better", "big", "blunt", "bold", "boss", "brave", "brief", "bright", "brisk", "busy", "calm", "canny", "caring", "casual", "causal", "chief", "choice", "civil", "clean", "clear", "clever", "close", "comic", "cool", "cosmic", "crack", "crisp", "cross", "cuddly", "cute", "daring", "dear", "decent", "deep", "direct", "divine", "dodgy", "driven", "eager", "easy", "enough", "epic", "equal", "exact", "exotic", "expert", "fair", "famous", "fancy", "fast", "fine", "finer", "firm", "first", "fishy", "fit", "fleet", "fluent", "flying", "fond", "frank", "free", "fresh", "full", "fun", "funky", "funny", "fuzzy", "game", "gentle", "giving", "glad", "gloomy", "golden", "good", "grand", "great", "grimy", "grown", "gruff", "guided", "guilty", "handy", "happy", "hardy", "helped", "heroic", "hip", "holy", "honest", "hot", "huge", "humane", "humble", "ideal", "immune", "in", "intent", "ironic", "joint", "just", "keen", "key", "kind", "known", "large", "legal", "lethal", "light", "liked", "live", "living", "loved", "lovely", "loving", "loyal", "lucky", "macho", "magic", "main", "major", "many", "master", "mature", "meet", "merry", "mighty", "mint", "model", "modern", "modest", "moody", "moral", "more", "moved", "moving", "mutual", "naive", "native", "nearby", "neat", "needed", "new", "next", "nice", "noble", "normal", "noted", "novel", "on", "one", "open", "pet", "picked", "pious", "poetic", "polite", "pretty", "prime", "pro", "prompt", "proper", "proud", "proven", "pumped", "pure", "quick", "quiet", "rapid", "rare", "ready", "real", "rested", "rich", "right", "robust", "rowdy", "ruling", "sacred", "safe", "saved", "saving", "secure", "select", "set", "sharp", "silly", "simple", "slick", "small", "smart", "smooth", "social", "solid", "sought", "sound", "square", "stable", "star", "steady", "stern", "sticky", "still", "strong", "sturdy", "subtle", "suited", "sunny", "super", "superb", "sure", "sweet", "swift", "tacky", "tender", "tidy", "tight", "top", "tops", "tough", "tragic", "true", "trusty", "unique", "united", "up", "upward", "usable", "useful", "valid", "valued", "vast", "viable", "vital", "vocal", "wanted", "warm", "weird", "well", "whole", "wired", "wise", "witty", "worthy"}

var animals = []string{"aardvark", "adder", "airedale", "akita", "albacore", "alien", "alpaca", "amoeba", "anchovy", "anemone", "ant", "anteater", "antelope", "ape", "aphid", "arachnid", "asp", "baboon", "badger", "barnacle", "basilisk", "bass", "bat", "beagle", "bear", "bedbug", "bee", "beetle", "bengal", "bird", "bison", "blowfish", "bluebird", "bluegill", "bluejay", "boa", "boar", "bobcat", "bonefish", "boxer", "bream", "buck", "buffalo", "bug", "bull", "bulldog", "bullfrog", "bunny", "burro", "buzzard", "caiman", "calf", "camel", "cardinal", "caribou", "cat", "catfish", "cattle", "chamois", "cheetah", "chicken", "chigger", "chimp", "chipmunk", "chow", "cicada", "civet", "clam", "cobra", "cockatoo", "cod", "collie", "colt", "condor", "coral", "corgi", "cougar", "cow", "cowbird", "coyote", "crab", "crane", "crappie", "crawdad", "crayfish", "cricket", "crow", "cub", "dane", "dassie", "deer", "dingo", "dinosaur", "doberman", "dodo", "doe", "dog", "dogfish", "dolphin", "donkey", "dory", "dove", "dragon", "drake", "drum", "duck", "duckling", "eagle", "earwig", "eel", "eft", "egret", "elephant", "elf", "elk", "emu", "escargot", "ewe", "falcon", "fawn", "feline", "ferret", "filly", "finch", "firefly", "fish", "flamingo", "flea", "flounder", "fly", "foal", "fowl", "fox", "foxhound", "frog", "gannet", "gar", "garfish", "gator", "gazelle", "gecko", "gelding", "ghost", "ghoul", "gibbon", "giraffe", "glider", "glowworm", "gnat", "gnu", "goat", "gobbler", "goblin", "goldfish", "goose", "gopher", "gorilla", "goshawk", "grackle", "griffon", "grizzly", "grouper", "grouse", "grub", "grubworm", "guinea", "gull", "guppy", "haddock", "hagfish", "halibut", "hamster", "hare", "hawk", "hedgehog", "hen", "hermit", "heron", "herring", "hippo", "hog", "honeybee", "hookworm", "hornet", "horse", "hound", "humpback", "husky", "hyena", "ibex", "iguana", "imp", "impala", "insect", "jackal", "jackass", "jaguar", "javelin", "jawfish", "jay", "jaybird", "jennet", "joey", "kangaroo", "katydid", "kid", "killdeer", "kingfish", "kit", "kite", "kitten", "kiwi", "koala", "kodiak", "koi", "krill", "lab", "labrador", "lacewing", "ladybird", "ladybug", "lamb", "lamprey", "lark", "leech", "lemming", "lemur", "leopard", "liger", "lion", "lioness", "lionfish", "lizard", "llama", "lobster", "locust", "longhorn", "loon", "louse", "lynx", "macaque", "macaw", "mackerel", "maggot", "magpie", "mako", "malamute", "mallard", "mammal", "mammoth", "man", "manatee", "mantis", "marlin", "marmoset", "marmot", "marten", "martin", "mastiff", "mastodon", "mayfly", "meerkat", "midge", "mink", "minnow", "mite", "moccasin", "mole", "mollusk", "molly", "monarch", "mongoose", "mongrel", "monitor", "monkey", "monkfish", "monster", "moose", "moray", "mosquito", "moth", "mouse", "mudfish", "mule", "mullet", "muskox", "muskrat", "mustang", "mutt", "narwhal", "newt", "oarfish", "ocelot", "octopus", "opossum", "orca", "oriole", "oryx", "osprey", "ostrich", "owl", "ox", "oyster", "panda", "pangolin", "panther", "parakeet", "parrot", "peacock", "pegasus", "pelican", "penguin", "perch", "pheasant", "phoenix", "pig", "pigeon", "piglet", "pika", "pipefish", "piranha", "platypus", "polecat", "polliwog", "pony", "poodle", "porpoise", "possum", "prawn", "primate", "pug", "puma", "pup", "python", "quagga", "quail", "quetzal", "rabbit", "raccoon", "racer", "ram", "raptor", "rat", "rattler", "raven", "ray", "redbird", "redfish", "reindeer", "reptile", "rhino", "ringtail", "robin", "rodent", "rooster", "roughy", "sailfish", "salmon", "satyr", "sawfish", "sawfly", "scorpion", "sculpin", "seagull", "seahorse", "seal", "seasnail", "serval", "shad", "shark", "sheep", "sheepdog", "shepherd", "shiner", "shrew", "shrimp", "silkworm", "skink", "skunk", "skylark", "sloth", "slug", "snail", "snake", "snapper", "snipe", "sole", "spaniel", "sparrow", "spider", "sponge", "squid", "squirrel", "stag", "stallion", "starfish", "starling", "stingray", "stinkbug", "stork", "stud", "sturgeon", "sunbeam", "sunbird", "sunfish", "swan", "swift", "swine", "tadpole", "tahr", "tapir", "tarpon", "teal", "termite", "terrapin", "terrier", "tetra", "thrush", "tick", "tiger", "titmouse", "toad", "tomcat", "tortoise", "toucan", "treefrog", "troll", "trout", "tuna", "turkey", "turtle", "unicorn", "urchin", "vervet", "viper", "vulture", "wahoo", "wallaby", "walleye", "walrus", "warthog", "wasp", "weasel", "weevil", "werewolf", "whale", "whippet", "wildcat", "wolf", "wombat", "woodcock", "worm", "wren", "yak", "yeti", "zebra"}

// generateUniqueEnvironmentName will generate a unique fiab-style name for a BEE. eg "bee-funky-hermit"
// (this algorithm is a dumb hack intended to last until we can do this better in Sherlock)
func generateUniqueEnvironmentName(prefix string, stateFile StateFile) (string, error) {
	if len(prefix) == 0 {
		prefix = defaultPrefix
	}

	for i := 0; i < maxUniqueNameAttempts; i++ {
		adjective := randomElement(adjectives)
		animal := randomElement(animals)
		name := strings.Join([]string{prefix, adjective, animal}, separator)

		if _, exists := stateFile.Environments[name]; !exists {
			log.Debug().Msgf("Generated unique name %s (%d previous attempts)", name, i)
			return name, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique environment name after %d attempts", maxUniqueNameAttempts)
}

func randomElement(list []string) string {
	index := rand.Intn(len(list))
	return list[index]
}
