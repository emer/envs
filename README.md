# envs

This repository contains various `emer.Env` environments for simulations, or ingredients for them, providing example code that can be adapted as needed.

Many examples are configured as a `main` package with a minimal GUI-based interactive testing program, so you can directly run and test the code without a full associated simulation.

# Strategies

## Separate the "script" from the "production"

For more complex environments, an approach analogous to making a movie can be productive, where you first generate a "script" (screenplay) for the events, and then have a separable system for turning that into a fully-rendered environment.  This separates the different levels of logic and constraints, and helps manage the complexity.

For example, the `saccade` package generates a script based on an abstract 2D visual environment, without any details of the objects involved.  It coordinates the timing of planning vs. execution, and updating current vs. next variables, etc, and renders everything as normalized coordinates in a table.  The `obj3d` package manages lists of 3D object files to present, with shuffled order and train / test sets.  We then use these two packages to animate 3D objects in the `eve` based `obj3dsac` model to provide movie-like inputs for a predictive learning model.

The `emer/esg` package can be a useful tool for generating randomly-sampled story-like inputs with various sources of syntactic and semantic structure -- it is used for example in the `sg` sentence-gestalt model in https://github.com/CompCogNeuro/sims `ch9/sg` to generate sentences with these structures.

