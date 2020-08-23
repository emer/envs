# obj3dsac

Obj3DSac generates renderings of 3D objects with saccadic eye movements.  Object is moved around using Sac positions relative to 0,0,0 origin and Camera is at 0,0,Z with rotations based on saccade movements.  Can save a table of saccade data plus image file names to use for offline training of models on a cluster using this code, or incorporate into an env for more dynamic uses.

# Files

This [google drive folder](https://drive.google.com/drive/folders/13Mi9aUlF1A3sx3JaofX-qzKlxGoViT86?usp=sharing) has rendered output and .obj input files for use in this env.

* `CU3D100_20obj_8tick4sac.tar` is 1000 epochs of training and 50 of testing, for 8 ticks, 4 saccades (every 2 ticks) with the Objs20 list, from O’Reilly, R. C., Russin, J. L., Zolfaghar, M., & Rohrlich, J. (2020). Deep Predictive Learning in Neocortex and Pulvinar. ArXiv:2006.14800 [q-Bio]. http://arxiv.org/abs/2006.14800

* forthcoming: .tar file of 3D obj files (being converted from sketchup originals)

