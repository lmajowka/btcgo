const address1 = "HsMJxNiV7TLxmoF6uJNkydxPFDog4NQum"
const address2 = "Jp6T3JBrBPD5hKEZtqBjbdxigZvh99ceE"

let result;

if (address1 < address2) {
    result = "menor";
} else if (address1 > address2) {
    result = "maior";
} else {
    result = "igual";
}

console.log(`O endereço "${address1}" é ${result} que o endereço "${address2}".`);
