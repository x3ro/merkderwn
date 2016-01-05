 The subpage size also becomes the minimum buffer size for read and write operations. It is possible to use subpage sizes smaller than $\frac{p}{k}$, with the result that part of the flash memory (<!--$\frac{s \cdot k}{p}$-->) becomes inaccessible because it can never be written to.

 This is expected behavior because of the way multimarkdown handles dollar signs: ($<!--\frac{s \cdot k}{p}-->$)

 This is a workaround: (<!--$\frac{s \cdot k}{p}$-->)
