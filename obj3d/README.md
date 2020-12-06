# obj3d

The `obj3d` package manages lists of 3D object files to present, with shuffled order and train / test sets. 

It looks for Wavefront .obj files in subdirectories organized by category names -- these .obj files can be loaded into a gi3d / eve scene for rendering.

This example is useful for seeing how to generate lists of directories and files.

# Files

This [google drive folder](https://drive.google.com/drive/folders/13Mi9aUlF1A3sx3JaofX-qzKlxGoViT86?usp=sharing) has .obj input files for use in this env.

* `CU3D_100_plus_models_obj.tar.gz` is a set of 100 3D object categories, with 14.45 average different instances per category, used for much of our object recognition work, including the above paper and originally: O'Reilly, R.C., Wyatte, D., Herd, S., Mingus, B. & Jilk, D.J. (2013). Recurrent Processing during Object Recognition. *Frontiers in Psychology, 4,* 124. [PDF](https://ccnlab.org/papers/OReillyWyatteHerdEtAl13.pdf) | [URL](http://www.ncbi.nlm.nih.gov/pubmed/23554596)

