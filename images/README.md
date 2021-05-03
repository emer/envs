# images

The `images` package manages lists of image files to present, with shuffled order and train / test sets. 

It looks for .jpg or .png files with filenames indicating object category, or in subdirectories organized by category names.

This example is useful for seeing how to generate lists of directories and files.

# Files

This [google drive folder](https://drive.google.com/drive/folders/13Mi9aUlF1A3sx3JaofX-qzKlxGoViT86?usp=sharing) has .png input files for use in this env.

`CU3D_100_plus_renders.tar.gz` is a set of 30,240 rendered images from 100 3D object categories, with 14.45 average different instances per category, used for much of our object recognition work, including the above paper and originally: O'Reilly, R.C., Wyatte, D., Herd, S., Mingus, B. & Jilk, D.J. (2013). Recurrent Processing during Object Recognition. *Frontiers in Psychology, 4,* 124. [PDF](https://ccnlab.org/papers/OReillyWyatteHerdEtAl13.pdf) | [URL](http://www.ncbi.nlm.nih.gov/pubmed/23554596)

The image specs are: 320x320 color images. 100 object classes, 20 images per exemplar. Rendered with 40° depth rotation about y-axis (plus horizontal flip), 20° tilt rotation about x-axis, 80° overhead lighting rotation. There are no affine in-plane transformations (translation, scale, rotation) because these are easily done on the fly as part of "data augmentation".

