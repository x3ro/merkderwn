 The subpage size also becomes the minimum buffer size for read and write operations. It is possible to use subpage sizes smaller than $\frac{p}{k}$, with the result that part of the flash memory (<!--$\frac{s \cdot k}{p}$-->) becomes inaccessible because it can never be written to.

 This is expected behavior because of the way multimarkdown handles dollar signs: ($<!--\frac{s \cdot k}{p}-->$)

 This is a workaround: (<!--$\frac{s \cdot k}{p}$-->)

In comparison, a recent survey puts the energy consumption when reading from a 8 GB flash chip at <!--$~0.001 \mu J$--> per Byte, and at <!--$~0.025 \mu J$--> when writing a Byte.
