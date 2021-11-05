# obj3dsac

Obj3DSac generates renderings of 3D objects with saccadic eye movements.  Object is moved around using Sac positions relative to 0,0,0 origin and Camera is at 0,0,Z with rotations based on saccade movements.  Can save a table of saccade data plus image file names to use for offline training of models on a cluster using this code, or incorporate into an env for more dynamic uses.

# Files

This [google drive folder](https://drive.google.com/drive/folders/13Mi9aUlF1A3sx3JaofX-qzKlxGoViT86?usp=sharing) has rendered output and .obj input files for use in this env.

* `CU3D_100_20obj8inst_8tick4sac.tar` is 1000 epochs of training and 50 of testing, for 8 ticks, 4 saccades (every 2 ticks) with the Objs20 list, from O'Reilly, R. C., Russin, J. L., Zolfaghar, M., & Rohrlich, J. (2021). Deep predictive learning in neocortex and pulvinar. *Journal of Cognitive Neuroscience, 33(6),* 1158–1196., http://arxiv.org/abs/2006.14800

* `CU3D_100_plus_models_obj.tar.gz` is a set of 100 3D object categories, with 14.45 average different instances per category, used for much of our object recognition work, including the above paper and originally: O'Reilly, R.C., Wyatte, D., Herd, S., Mingus, B. & Jilk, D.J. (2013). Recurrent Processing during Object Recognition. *Frontiers in Psychology, 4,* 124. [PDF](https://ccnlab.org/papers/OReillyWyatteHerdEtAl13.pdf) | [URL](http://www.ncbi.nlm.nih.gov/pubmed/23554596)




