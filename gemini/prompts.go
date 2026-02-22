package gemini

const detectCategoryPromptTemplate = `Classify the German word into one category: noun, verb, or other.
Word: %s
Return only the category word.`

const defaultPromptTemplate = `
Return the following fields in a JSON structure for the word: %s
The values will be used for creating Anki cards to learn German vocabulary.

* base_d: the base form the German word.
  * When a noun, omit the article. e.g. "Abgas".
	* When a reflexive verb, start with "sich".
* full_d: German word. 
  * When a verb, should be a comma separated list of infinitive, present, simple past, and present perfect. e.g. "analysieren, analysiert, analysierte, hat analysiert"
	* When a reflexive verb, start with "sich". e.g. "sich amüsieren, amüsiert sich, amüsierte sich, hat sich amüsiert"
  * When a noun, should include the article, and the ending in plural. e.g. "das Abgas, -e", "das Alter, -". This is just a combination of the fields artikel_d, base_d, and plural_d.
* base_e: the English translation. e.g. "to analyze"
  * If an English translation is provided in the prompt, make sure base_e covers what is provided
* artikel_d:
  * When a noun, the article. 
  * When not a noun, blank string
* plural_d: 
  * When a noun, the plural ending. "-" if the ending does not change, and e.g. "-e" if an "e" is added.
  * When not a noun, blank string
* s1: first example sentence in German. Create a typical sentence that the word would be used in.
* s1e: english translation of s1
* s2: second example sentence in German
  * Only include if all commonly used meanings haven't been covered by previous examples. Otherwise, leave blank.
* s2e: english translation of s2
* s3: third example sentence in German
  * Only include if all commonly used meanings haven't been covered by previous examples. Otherwise, leave blank.
* s3e: english translation of s3
* s4: fourth example sentence in German
  * Only include if all commonly used meanings haven't been covered by previous examples. Otherwise, leave blank.
* s4e: english translation of s4

Other things to note: 
* If the word is in plural, convert it to singular
* Return ONLY the JSON object wrapped in a json code block, and do not include any other content or text.
`

const nounPromptTemplate = `
Return the following fields in a JSON structure for the noun: %s
This is used for creating Anki cards to learn German vocabulary.

* base_d: the base form the German word. Omit the article.
* full_d: German noun with article and full plural form. e.g. "das Abgas, die Abgase", "das Alter, die Alter"
* base_e: the English translation/meanings. e.g. "exhaust gas"
  * if there area multiple meanings, separate different meanings with a ";", and group similar meanings with "/".
* artikel_d: the article.
* plural_d: the full plural form (include the article and plural word). e.g. "die Abgase", "die Alter"
* s1: 1st example sentence in German
  * Create a typical sentence that starts with the article and noun in nominative.
* s1e: english translation of s1
* s2: 2nd example sentence in German
  * Create a typical sentence that includes the plural of the noun.
	* If the plural of the noun is not used commonly, demonstrate another of the words meanings.
	* If there are no other meanings, just create a different sentence.
* s2e: english translation of s2
* s3: 3rd example sentence in German
  * Only include if all commonly used meanings haven't been covered by previous examples. Otherwise, leave blank.
* s3e: english translation of s3
* s4: 4th example sentence in German
  * Only include if all commonly used meanings haven't been covered by previous examples. Otherwise, leave blank.
* s4e: english translation of s4

Other things to note:
* If the word is in plural, convert it to singular
* German and English sentences should end with appropriate punctuation (typically ".")
* Return ONLY the JSON object wrapped in a json code block, and do not include any other content or text.
`

const verbPromptTemplate = `
Return the following fields in a JSON structure for the verb: %s
The values will be used for creating Anki cards to learn German vocabulary.

* base_d: the base form of the German verb. 
  * if the verb is always reflexive, start with "sich".
* full_d: German verb as a comma separated list of infinitive, 3rd person present, 3rd person simple past, and past participle.
  * e.g. "analysieren, analysiert, analysierte, analysiert"
  * if the verb is always reflexive, start with "sich". e.g. "sich amüsieren, amüsiert sich, amüsierte sich, sich amüsiert"
* base_e: the English translation. e.g. "to analyze"
  * always start with "to"
* artikel_d: blank string
* plural_d: blank string
* s1: The first example sentence in German.
  * Create a typical sentence that the word would be used in, with the base form. 
* s1e: english translation of s1
* s2: 2nd example sentence in German
  * Create a typical sentence that the word would be used in, with the present perfect.
* s2e: english translation of s2
* s3: 3rd example sentence in German
  * Only include if all commonly used meanings haven't been covered by previous examples. Otherwise, leave blank.
* s3e: english translation of s3
* s4: 4th example sentence in German
  * Only include if all commonly used meanings haven't been covered by previous examples. Otherwise, leave blank.
* s4e: english translation of s4

Other things to note:
* German and English sentences should end with appropriate punctuation (typically ".")
* Return ONLY the JSON object wrapped in a json code block, and do not include any other content or text.
`
