CorpusEnv reads text from a file and presents a window of sequential words as the input.  Words included in the vocabulary can be filtered by frequency at both ends.

For input, a simple bag-of-words is used, with words encoded as localist 1-hot units (requiring a large input layer) or using random distributed vectors in a lower-dimensional space.

Contributed by Taiqi He and Randy O'Reilly

