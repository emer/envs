# Approach

This environment is a simplified version of [fworld](https://github.com/emer/envs/tree/master/fworld) focused on approaching things you want / need.  Target behavior is to orient L / R until a CS sensory cue appears that is consistent with current body state, and then move Forward until the Distance = proximal, and you then Consume.

* Actions are: Forward, Left, Right, Consume: move forward to approach desired item, go L or R to look for something you desire.  Consume when you get to zero distance -- no effect before then (waste of time).

* `Drives`: different drive-like body states (hunger, thirst, etc), that are satisfied by a corresponding US outcome.

* `CSPerDrive`: number of different CS sensory cues associated with each US (simplest case is 1 -- one-to-one mapping), presented on a "fovea" input layer.

* `Locactions`: different locations where you can be currently looking, each of which can hold a different CS sensory cue and associated US, available when you approach and consume.  Current location is an input, and determines contents of the fovea.  wraps around.

* `DistMax`: Maximum Distance in time steps -- actual starting distance can be less than this.  This can be represented as a popcode.  Start at random distances.




